package api

import (
	"context"
	"net/http"

	"github.com/coreos/go-oidc"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/oauth2"

	"github.com/doutorfinancas/pun-sho/api/request"
	"github.com/doutorfinancas/pun-sho/service"
)

type AuthHandler struct {
	log               *zap.Logger
	authSvc           *service.AuthService
	cookieDomain      string
	localLoginEnabled bool
	msOAuthConfig     *oauth2.Config
	msOIDCVerifier    *oidc.IDTokenVerifier
	allowedGroups     []string
}

func NewAuthHandler(
	log *zap.Logger,
	authSvc *service.AuthService,
	cookieDomain string,
	disableLocalLogin bool,
) *AuthHandler {
	return &AuthHandler{
		log:               log,
		authSvc:           authSvc,
		cookieDomain:      cookieDomain,
		localLoginEnabled: !disableLocalLogin,
	}
}

func (h *AuthHandler) ConfigureMicrosoftOAuth(tenantID, clientID, clientSecret string, allowedGroups []string) {
	if tenantID == "" || clientID == "" || clientSecret == "" {
		return
	}

	ctx := context.Background()
	provider, err := oidc.NewProvider(ctx, "https://login.microsoftonline.com/"+tenantID+"/v2.0")
	if err != nil {
		h.log.Error("Failed to initialize MS OIDC provider", zap.Error(err))
		return
	}

	h.msOIDCVerifier = provider.Verifier(&oidc.Config{ClientID: clientID})
	h.msOAuthConfig = &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     provider.Endpoint(),
		RedirectURL:  "", // Set dynamically based on request
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}
	h.allowedGroups = allowedGroups

	h.log.Info("Microsoft OAuth configured", zap.String("tenant", tenantID))
}

func (h *AuthHandler) Validate() {
	if !h.localLoginEnabled && h.msOAuthConfig == nil {
		h.log.Warn("DISABLE_LOCAL_LOGIN is set but Microsoft OAuth is not configured, re-enabling local login")
		h.localLoginEnabled = true
	}
}

func (h *AuthHandler) Routes(rg *gin.RouterGroup) {
	rg.GET("/login", h.loginPage)
	rg.POST("/logout", h.logout)

	if h.localLoginEnabled {
		rg.POST("/login", h.loginAction)
		rg.POST("/login/totp", h.totpVerify)
	}

	if h.msOAuthConfig != nil {
		rg.GET("/auth/ms", h.msRedirect)
		rg.GET("/auth/ms/callback", h.msCallback)
	}
}

func (h *AuthHandler) Group() *string {
	return nil
}

func (h *AuthHandler) loginPage(c *gin.Context) {
	msEnabled := h.msOAuthConfig != nil

	// If local login is disabled and MS is enabled, redirect directly to MS
	if !h.localLoginEnabled && msEnabled && c.Query("error") == "" {
		h.msRedirect(c)
		return
	}

	data := gin.H{
		"MicrosoftEnabled":  msEnabled,
		"LocalLoginEnabled": h.localLoginEnabled,
	}

	if errMsg := c.Query("error"); errMsg != "" {
		data["Error"] = errMsg
	}

	c.Header("Content-Type", "text/html; charset=utf-8")
	renderTemplate(c, "login.html", data)
}

func (h *AuthHandler) loginAction(c *gin.Context) {
	var req request.Login
	if err := c.ShouldBind(&req); err != nil {
		c.Redirect(http.StatusFound, "/app/login?error=Invalid+request")
		return
	}

	user, err := h.authSvc.Login(req.Username, req.Password)
	if err != nil {
		c.Redirect(http.StatusFound, "/app/login?error=Invalid+credentials")
		return
	}

	if user.TOTPEnabled {
		// Create a pending session token for TOTP verification
		session, err := h.authSvc.CreateSession(user.ID)
		if err != nil {
			c.Redirect(http.StatusFound, "/app/login?error=Session+error")
			return
		}

		data := gin.H{
			"TOTPRequired":     true,
			"PendingToken":     session.Token,
			"MicrosoftEnabled": h.msOAuthConfig != nil,
			"LocalLoginEnabled": h.localLoginEnabled,
		}
		c.Header("Content-Type", "text/html; charset=utf-8")
		renderTemplate(c, "login.html", data)
		return
	}

	h.createSessionAndRedirect(c, user.ID)
}

func (h *AuthHandler) totpVerify(c *gin.Context) {
	var req request.TOTPVerify
	if err := c.ShouldBind(&req); err != nil {
		c.Redirect(http.StatusFound, "/app/login?error=Invalid+request")
		return
	}

	user, err := h.authSvc.ValidateSession(req.SessionToken)
	if err != nil {
		c.Redirect(http.StatusFound, "/app/login?error=Session+expired")
		return
	}

	if !h.authSvc.ValidateTOTP(user, req.Code) {
		// Delete the pending session
		_ = h.authSvc.Logout(req.SessionToken)
		c.Redirect(http.StatusFound, "/app/login?error=Invalid+TOTP+code")
		return
	}

	// TOTP valid — set the cookie with the existing session token
	c.SetCookie(SessionCookieName, req.SessionToken, 60*60*48, "/", h.cookieDomain, false, true)
	c.Redirect(http.StatusFound, "/app/")
}

func (h *AuthHandler) logout(c *gin.Context) {
	token, _ := c.Cookie(SessionCookieName)
	if token != "" {
		_ = h.authSvc.Logout(token)
	}
	c.SetCookie(SessionCookieName, "", -1, "/", h.cookieDomain, false, true)
	c.Redirect(http.StatusFound, "/app/login")
}

func (h *AuthHandler) msRedirect(c *gin.Context) {
	if h.msOAuthConfig == nil {
		c.Redirect(http.StatusFound, "/app/login")
		return
	}

	config := *h.msOAuthConfig
	config.RedirectURL = scheme(c) + "://" + c.Request.Host + "/app/auth/ms/callback"

	url := config.AuthCodeURL("state")
	c.Redirect(http.StatusFound, url)
}

func (h *AuthHandler) msCallback(c *gin.Context) {
	if h.msOAuthConfig == nil {
		c.Redirect(http.StatusFound, "/app/login")
		return
	}

	code := c.Query("code")
	if code == "" {
		c.Redirect(http.StatusFound, "/app/login?error=No+authorization+code")
		return
	}

	config := *h.msOAuthConfig
	config.RedirectURL = scheme(c) + "://" + c.Request.Host + "/app/auth/ms/callback"

	ctx := c.Request.Context()
	token, err := config.Exchange(ctx, code)
	if err != nil {
		h.log.Error("MS OAuth exchange failed", zap.Error(err))
		c.Redirect(http.StatusFound, "/app/login?error=Authentication+failed")
		return
	}

	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		c.Redirect(http.StatusFound, "/app/login?error=Missing+ID+token")
		return
	}

	idToken, err := h.msOIDCVerifier.Verify(ctx, rawIDToken)
	if err != nil {
		c.Redirect(http.StatusFound, "/app/login?error=Invalid+ID+token")
		return
	}

	var claims struct {
		Email             string `json:"email"`
		PreferredUsername string `json:"preferred_username"`
	}
	if err := idToken.Claims(&claims); err != nil {
		c.Redirect(http.StatusFound, "/app/login?error=Failed+to+parse+claims")
		return
	}

	email := claims.Email
	if email == "" {
		email = claims.PreferredUsername
	}

	user, err := h.authSvc.FindOrCreateMSUser(email, email)
	if err != nil {
		c.Redirect(http.StatusFound, "/app/login?error=Account+creation+failed")
		return
	}

	h.createSessionAndRedirect(c, user.ID)
}

func (h *AuthHandler) createSessionAndRedirect(c *gin.Context, userID uuid.UUID) {
	session, err := h.authSvc.CreateSession(userID)
	if err != nil {
		c.Redirect(http.StatusFound, "/app/login?error=Session+error")
		return
	}
	c.SetCookie(SessionCookieName, session.Token, 60*60*48, "/", h.cookieDomain, false, true)
	c.Redirect(http.StatusFound, "/app/")
}

func scheme(c *gin.Context) string {
	if c.Request.TLS != nil {
		return "https"
	}
	if proto := c.GetHeader("X-Forwarded-Proto"); proto != "" {
		return proto
	}
	return "http"
}

