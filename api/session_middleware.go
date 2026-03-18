package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/doutorfinancas/pun-sho/service"
)

const (
	SessionCookieName = "pun_sho_session"
	ContextUserKey    = "user"
	ContextUserIDKey  = "user_id"
	ContextRoleKey    = "user_role"
)

type SessionMiddleware struct {
	authSvc *service.AuthService
}

func NewSessionMiddleware(authSvc *service.AuthService) *SessionMiddleware {
	return &SessionMiddleware{authSvc: authSvc}
}

func (m *SessionMiddleware) RequireSession(c *gin.Context) {
	token, err := c.Cookie(SessionCookieName)
	if err != nil || token == "" {
		c.Redirect(http.StatusFound, "/app/login")
		c.Abort()
		return
	}

	user, err := m.authSvc.ValidateSession(token)
	if err != nil {
		c.SetCookie(SessionCookieName, "", -1, "/", "", false, true)
		c.Redirect(http.StatusFound, "/app/login")
		c.Abort()
		return
	}

	c.Set(ContextUserKey, user)
	c.Set(ContextUserIDKey, user.ID.String())
	c.Set(ContextRoleKey, user.Role)
	c.Next()
}

func (m *SessionMiddleware) RequireAdmin(c *gin.Context) {
	role, exists := c.Get(ContextRoleKey)
	if !exists || role != "admin" {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}
	c.Next()
}
