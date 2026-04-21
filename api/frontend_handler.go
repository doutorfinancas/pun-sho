package api

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	htmlpkg "html"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/doutorfinancas/pun-sho/api/request"
	"github.com/doutorfinancas/pun-sho/entity"
	"github.com/doutorfinancas/pun-sho/service"
)

const maxLogoBytes = 1 << 20 // 1 MB

type FrontendHandler struct {
	log          *zap.Logger
	shortySvc    *service.ShortyService
	analyticsSvc *service.AnalyticsService
	authSvc      *service.AuthService
	hostName     string
}

func NewFrontendHandler(
	log *zap.Logger,
	shortySvc *service.ShortyService,
	analyticsSvc *service.AnalyticsService,
	authSvc *service.AuthService,
	hostName string,
) *FrontendHandler {
	return &FrontendHandler{
		log:          log,
		shortySvc:    shortySvc,
		analyticsSvc: analyticsSvc,
		authSvc:      authSvc,
		hostName:     hostName,
	}
}

func (h *FrontendHandler) Routes(rg *gin.RouterGroup) {
	// Full page routes
	rg.GET("/", h.dashboard)
	rg.GET("/links", h.linksList)
	rg.GET("/links/new", h.linksNewForm)
	rg.GET("/links/:id", h.linkDetail)
	rg.GET("/links/:id/edit", h.linkEditForm)
	rg.POST("/links", h.linksCreate)
	rg.POST("/links/:id", h.linksUpdate)
	rg.GET("/users", h.usersPage)

	// HTMX partial endpoints
	htmx := rg.Group("/htmx")
	htmx.GET("/dashboard/stats", h.htmxDashboardStats)
	htmx.GET("/dashboard/chart", h.htmxDashboardChart)
	htmx.GET("/dashboard/labels", h.htmxDashboardLabels)
	htmx.GET("/dashboard/recent", h.htmxDashboardRecent)
	htmx.GET("/links/list", h.htmxLinksList)
	htmx.GET("/links/check-slug", h.htmxCheckSlug)
	htmx.DELETE("/links/:id", h.htmxDeleteLink)
	htmx.GET("/links/:id/chart", h.htmxLinkChart)
	htmx.GET("/links/:id/browsers", h.htmxLinkBrowsers)
	htmx.GET("/links/:id/os", h.htmxLinkOS)
	htmx.GET("/links/:id/referrers", h.htmxLinkReferrers)
	htmx.GET("/links/:id/geo", h.htmxLinkGeo)
	htmx.GET("/links/:id/visits", h.htmxLinkVisits)
	htmx.POST("/users", h.htmxCreateUser)
	htmx.PATCH("/users/:id", h.htmxUpdateUser)
	htmx.DELETE("/users/:id", h.htmxDeleteUser)
}

func (h *FrontendHandler) Group() *string {
	return nil
}

func (h *FrontendHandler) baseData(c *gin.Context, activePage, pageTitle string) gin.H {
	role, _ := c.Get(ContextRoleKey)
	user, _ := c.Get(ContextUserKey)
	username := ""
	if u, ok := user.(*entity.User); ok {
		username = u.Username
	}

	return gin.H{
		"ActivePage": activePage,
		"PageTitle":  pageTitle,
		"UserRole":   role,
		"Username":   username,
		"HostName":   h.hostName,
	}
}

// --- Full Page Handlers ---

func (h *FrontendHandler) dashboard(c *gin.Context) {
	data := h.baseData(c, "dashboard", "Dashboard")
	renderTemplate(c, "layout.html", data)
}

func (h *FrontendHandler) linksList(c *gin.Context) {
	data := h.baseData(c, "links", "Links")

	availableLabels, err := h.shortySvc.AllLabels()
	if err != nil {
		availableLabels = []string{}
	}
	data["AvailableLabels"] = availableLabels

	renderTemplate(c, "layout.html", data)
}

func (h *FrontendHandler) linksNewForm(c *gin.Context) {
	data := h.baseData(c, "link_form", "Create Link")
	data["Labels"] = []string{}
	renderTemplate(c, "layout.html", data)
}

func (h *FrontendHandler) linkDetail(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Redirect(http.StatusFound, "/app/links")
		return
	}

	link, err := h.shortySvc.FindShortyByID(id, "", "", true)
	if err != nil {
		c.Redirect(http.StatusFound, "/app/links")
		return
	}
	link.ShortLink = fmt.Sprintf("%s/s/%s", h.hostName, link.PublicID)

	data := h.baseData(c, "link_detail", "Link Details")
	data["Link"] = link
	renderTemplate(c, "layout.html", data)
}

func (h *FrontendHandler) linkEditForm(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Redirect(http.StatusFound, "/app/links")
		return
	}

	link, err := h.shortySvc.FindShortyByID(id, "", "", false)
	if err != nil {
		c.Redirect(http.StatusFound, "/app/links")
		return
	}

	data := h.baseData(c, "link_form", "Edit Link")
	data["Link"] = link
	data["Labels"] = link.Labels
	renderTemplate(c, "layout.html", data)
}

func (h *FrontendHandler) linksCreate(c *gin.Context) {
	req := request.CreateShorty{
		Link:   c.PostForm("link"),
		Labels: c.PostFormArray("labels"),
	}

	// Parse TTL — only set if non-empty
	if ttlStr := c.PostForm("ttl"); ttlStr != "" {
		t, err := time.ParseInLocation("2006-01-02T15:04", ttlStr, time.Local)
		if err == nil {
			req.TTL = &t
		}
	}

	// Parse redirection limit — only set if non-empty
	if limitStr := c.PostForm("redirection_limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			req.RedirectionLimit = &limit
		}
	}

	// Parse UTM from form fields
	utm := &request.UTMParams{
		Source:   c.PostForm("utm_source"),
		Medium:   c.PostForm("utm_medium"),
		Campaign: c.PostForm("utm_campaign"),
		Term:     c.PostForm("utm_term"),
		Content:  c.PostForm("utm_content"),
	}
	if !utm.IsEmpty() {
		req.UTM = utm
	}

	// Parse slug from form
	if slug := c.PostForm("slug"); slug != "" {
		req.Slug = &slug
	}

	// Handle QR code
	if c.PostForm("enable_qr") == "on" {
		req.QRCode = buildQRRequest(c)
	}

	_, err := h.shortySvc.Create(&req)
	if err != nil {
		data := h.baseData(c, "link_form", "Create Link")
		data["Error"] = err.Error()
		data["Labels"] = req.Labels
		renderTemplate(c, "layout.html", data)
		return
	}

	c.Redirect(http.StatusFound, "/app/links")
}

func (h *FrontendHandler) linksUpdate(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Redirect(http.StatusFound, "/app/links")
		return
	}

	link, err := h.shortySvc.FindShortyByID(id, "", "", false)
	if err != nil {
		c.Redirect(http.StatusFound, "/app/links")
		return
	}

	req := request.UpdateShorty{
		Link:   c.PostForm("link"),
		Labels: c.PostFormArray("labels"),
	}

	if ttlStr := c.PostForm("ttl"); ttlStr != "" {
		t, parseErr := time.ParseInLocation("2006-01-02T15:04", ttlStr, time.Local)
		if parseErr == nil {
			req.TTL = &t
		}
	}

	if limitStr := c.PostForm("redirection_limit"); limitStr != "" {
		if limit, parseErr := strconv.Atoi(limitStr); parseErr == nil && limit > 0 {
			req.RedirectionLimit = &limit
		}
	}

	// Parse UTM from form fields
	utm := &request.UTMParams{
		Source:   c.PostForm("utm_source"),
		Medium:   c.PostForm("utm_medium"),
		Campaign: c.PostForm("utm_campaign"),
		Term:     c.PostForm("utm_term"),
		Content:  c.PostForm("utm_content"),
	}
	if !utm.IsEmpty() {
		req.UTM = utm
	}

	// Handle QR code
	if c.PostForm("enable_qr") == "on" {
		qrReq := buildQRRequest(c)
		shortLink := fmt.Sprintf("%s/s/%s", h.hostName, link.PublicID)
		qrCode, qrErr := h.shortySvc.RegenerateQR(qrReq, shortLink)
		if qrErr == nil {
			link.QRCode = qrCode
		}
	} else if link.QRCode != "" {
		link.QRCode = ""
	}

	_, err = h.shortySvc.Update(&req, link)
	if err != nil {
		c.Redirect(http.StatusFound, fmt.Sprintf("/app/links/%s/edit", id))
		return
	}

	c.Redirect(http.StatusFound, fmt.Sprintf("/app/links/%s", id))
}

func (h *FrontendHandler) usersPage(c *gin.Context) {
	role, _ := c.Get(ContextRoleKey)
	if role != "admin" {
		c.Redirect(http.StatusFound, "/app/")
		return
	}

	users, err := h.authSvc.ListUsers()
	if err != nil {
		users = []*entity.User{}
	}

	data := h.baseData(c, "users", "User Management")
	data["Users"] = users
	renderTemplate(c, "layout.html", data)
}

// --- HTMX Partial Handlers ---

func (h *FrontendHandler) htmxDashboardStats(c *gin.Context) {
	until := time.Now()
	from := until.AddDate(0, 0, -30)

	stats := h.analyticsSvc.GlobalSummary(from, until)
	renderTemplate(c, "stats_cards", gin.H{
		"TotalLinks":  stats.TotalLinks,
		"TotalClicks": stats.TotalClicks,
		"ActiveLinks": stats.ActiveLinks,
		"ExpiredLinks": stats.ExpiredLinks,
	})
}

func (h *FrontendHandler) htmxDashboardChart(c *gin.Context) {
	granularity := c.DefaultQuery("granularity", "day")
	until := time.Now()
	from := until.AddDate(0, 0, -30)

	points := h.analyticsSvc.ClicksOverTime(nil, from, until, granularity)
	renderTemplate(c, "chart", gin.H{"Points": points})
}

func (h *FrontendHandler) htmxDashboardLabels(c *gin.Context) {
	until := time.Now()
	from := until.AddDate(0, 0, -30)

	labels := h.analyticsSvc.LabelRanking(from, until, 10)

	html := ""
	if len(labels) == 0 {
		html = `<p class="text-gray-500 text-center">No label data available</p>`
	} else {
		for _, l := range labels {
			html += fmt.Sprintf(`<div class="d-flex justify-content-between align-items-center mb-2">
				<span class="label-chip">%s</span>
				<span class="fw-semibold">%d</span>
			</div>`, htmlpkg.EscapeString(l.Label), l.Count)
		}
	}

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, html)
}

func (h *FrontendHandler) htmxDashboardRecent(c *gin.Context) {
	links, err := h.shortySvc.List(false, nil, "", nil, nil, 5, 0)
	if err != nil {
		c.String(http.StatusOK, `<p class="text-gray-500 text-center">No links found</p>`)
		return
	}

	html := `<div class="table-responsive"><table class="table table-sm mb-0"><tbody>`
	for _, l := range links {
		shortLink := fmt.Sprintf("%s/s/%s", h.hostName, l.PublicID)
		html += fmt.Sprintf(`<tr>
			<td><a href="/app/links/%s" class="text-decoration-none fw-semibold">%s</a></td>
			<td class="truncate" style="max-width:200px;">%s</td>
			<td class="text-end"><span class="fw-semibold">%d</span> clicks</td>
		</tr>`, l.ID, shortLink, l.Link, l.Visits)
	}
	html += `</tbody></table></div>`

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, html)
}

func (h *FrontendHandler) htmxLinksList(c *gin.Context) {
	labels := c.QueryArray("labels")
	status := c.Query("status")

	var from, to *time.Time
	if fromStr := c.Query("from"); fromStr != "" {
		if t, err := time.Parse("2006-01-02", fromStr); err == nil {
			from = &t
		}
	}
	if toStr := c.Query("to"); toStr != "" {
		if t, err := time.Parse("2006-01-02", toStr); err == nil {
			end := t.Add(24*time.Hour - time.Second)
			to = &end
		}
	}

	links, err := h.shortySvc.List(false, labels, status, from, to, 50, 0)
	if err != nil {
		c.String(http.StatusOK, `<tr><td colspan="7" class="text-center text-gray-500">Error loading links</td></tr>`)
		return
	}

	if len(links) == 0 {
		c.String(http.StatusOK, `<tr><td colspan="7" class="text-center py-4 text-gray-500">No links found. <a href="/app/links/new">Create one!</a></td></tr>`)
		return
	}

	html := ""
	for _, l := range links {
		shortLink := fmt.Sprintf("%s/s/%s", h.hostName, l.PublicID)
		statusBadge := `<span class="status-badge status-success">Active</span>`
		if l.DeletedAt != nil {
			statusBadge = `<span class="status-badge status-error">Deleted</span>`
		} else if l.TTL != nil && l.TTL.Before(time.Now()) {
			statusBadge = `<span class="status-badge status-warning">Expired</span>`
		}

		labels := ""
		for _, label := range l.Labels {
			labels += fmt.Sprintf(`<span class="label-chip">%s</span>`, htmlpkg.EscapeString(label))
		}

		created := ""
		if l.CreatedAt != nil {
			created = l.CreatedAt.Format("Jan 02, 2006")
		}

		deleteBtn := ""
		if l.DeletedAt == nil {
			deleteBtn = fmt.Sprintf(`<button class="btn btn-sm btn-outline-danger" hx-delete="/app/htmx/links/%s" hx-target="closest tr" hx-swap="outerHTML" hx-confirm="Delete this link?"><i class="df-icon-delete"></i></button>`, l.ID)
		}

		escapedLink := htmlpkg.EscapeString(l.Link)
		html += fmt.Sprintf(`<tr>
			<td class="ps-4"><div class="d-flex align-items-center"><a href="/app/links/%s" class="text-decoration-none fw-semibold">%s</a><button class="btn btn-sm btn-link p-0 ms-1 copy-btn" onclick="copyToClipboard('%s', this)"><i class="df-icon-copy"></i></button></div></td>
			<td><span class="truncate d-inline-block" style="max-width:300px;" data-bs-toggle="tooltip" title="%s">%s</span></td>
			<td><span class="fw-semibold">%d</span> <span class="text-gray-500 small">(%d redirects)</span></td>
			<td>%s</td>
			<td>%s</td>
			<td class="small text-gray-600">%s</td>
			<td class="pe-4"><div class="d-flex gap-1"><a href="/app/links/%s" class="btn btn-sm btn-outline-primary"><i class="df-icon-chart-data"></i></a><a href="/app/links/%s/edit" class="btn btn-sm btn-outline-secondary"><i class="df-icon-edit"></i></a>%s</div></td>
		</tr>`, l.ID, shortLink, shortLink, escapedLink, escapedLink, l.Visits, l.Redirects, statusBadge, labels, created, l.ID, l.ID, deleteBtn)
	}

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, html)
}

func (h *FrontendHandler) htmxCheckSlug(c *gin.Context) {
	slug := c.Query("slug")
	if slug == "" {
		c.String(http.StatusOK, "")
		return
	}

	if len(slug) < 3 {
		c.String(http.StatusOK, `<span class="text-warning"><i class="df-icon-warning"></i></span>`)
		return
	}

	exists, _ := h.shortySvc.FindShortyByPublicID(slug)
	if exists != nil {
		c.String(http.StatusOK, `<span class="text-danger"><i class="df-icon-cancel"></i></span>`)
		return
	}

	c.String(http.StatusOK, `<span class="text-success"><i class="df-icon-check"></i></span>`)
}

func (h *FrontendHandler) htmxDeleteLink(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	if err := h.shortySvc.DeleteShortyByUUID(id); err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	c.Header("HX-Toast-Success", "Link deleted successfully")
	c.String(http.StatusOK, "")
}

func (h *FrontendHandler) htmxLinkChart(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	until := time.Now()
	from := until.AddDate(0, 0, -30)
	granularity := c.DefaultQuery("granularity", "day")

	points := h.analyticsSvc.ClicksOverTime(&id, from, until, granularity)
	renderTemplate(c, "chart", gin.H{"Points": points})
}

func (h *FrontendHandler) htmxLinkBrowsers(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	until := time.Now()
	from := until.AddDate(0, 0, -30)

	items := h.analyticsSvc.BrowserBreakdown(&id, from, until)
	renderBreakdownTable(c, items, "Browser")
}

func (h *FrontendHandler) htmxLinkOS(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	until := time.Now()
	from := until.AddDate(0, 0, -30)

	items := h.analyticsSvc.OSBreakdown(&id, from, until)
	renderBreakdownTable(c, items, "Operating System")
}

func (h *FrontendHandler) htmxLinkReferrers(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	until := time.Now()
	from := until.AddDate(0, 0, -30)

	items := h.analyticsSvc.TopReferrers(&id, from, until, 10)
	renderBreakdownTable(c, items, "Referrer")
}

func (h *FrontendHandler) htmxLinkGeo(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	until := time.Now()
	from := until.AddDate(0, 0, -30)

	items := h.analyticsSvc.LocationBreakdown(&id, from, until, 10)
	if len(items) == 0 {
		c.String(http.StatusOK, `<p class="text-gray-500 text-center">No geographic data available</p>`)
		return
	}

	html := `<table class="table table-sm"><thead><tr><th>Country</th><th>City</th><th class="text-end">Clicks</th></tr></thead><tbody>`
	for _, item := range items {
		html += fmt.Sprintf(`<tr><td>%s</td><td>%s</td><td class="text-end fw-semibold">%d</td></tr>`, item.Country, item.City, item.Count)
	}
	html += `</tbody></table>`

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, html)
}

func (h *FrontendHandler) htmxLinkVisits(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	accesses := h.shortySvc.FindAllAccessesByShortyID(id)

	if len(accesses) == 0 {
		c.String(http.StatusOK, `<p class="text-gray-500 text-center">No visits recorded</p>`)
		return
	}

	html := `<div class="table-responsive"><table class="table table-sm"><thead><tr><th>Time</th><th>IP</th><th>Browser</th><th>OS</th><th>Status</th><th>Country</th></tr></thead><tbody>`
	limit := 50
	if len(accesses) < limit {
		limit = len(accesses)
	}
	for i := len(accesses) - 1; i >= len(accesses)-limit && i >= 0; i-- {
		a := accesses[i]
		statusClass := "status-success"
		if a.Status != "redirected" {
			statusClass = "status-warning"
		}
		created := ""
		if a.CreatedAt != nil {
			created = a.CreatedAt.Format("Jan 02 15:04")
		}
		html += fmt.Sprintf(`<tr><td class="small">%s</td><td class="small">%s</td><td class="small">%s</td><td class="small">%s</td><td><span class="status-badge %s">%s</span></td><td class="small">%s</td></tr>`,
			created, a.IPAddress, a.Browser, a.OperatingSystem, statusClass, a.Status, a.Country)
	}
	html += `</tbody></table></div>`

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, html)
}

// --- User Management HTMX Handlers ---

func (h *FrontendHandler) htmxCreateUser(c *gin.Context) {
	role, _ := c.Get(ContextRoleKey)
	if role != "admin" {
		c.Status(http.StatusForbidden)
		return
	}

	var req request.CreateUser
	if err := c.ShouldBind(&req); err != nil {
		c.String(http.StatusBadRequest, "Invalid form data")
		return
	}

	if req.Role == "" {
		req.Role = "user"
	}

	_, err := h.authSvc.CreateUser(req.Username, req.Email, req.Password, req.Role)
	if err != nil {
		c.String(http.StatusBadRequest, "Error creating user: %s", err.Error())
		return
	}

	c.Header("HX-Toast-Success", "User created successfully")
	h.renderUsersTableBody(c)
}

func (h *FrontendHandler) htmxUpdateUser(c *gin.Context) {
	role, _ := c.Get(ContextRoleKey)
	if role != "admin" {
		c.Status(http.StatusForbidden)
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	action := c.PostForm("action")
	toastMsg := "User updated successfully"
	switch action {
	case "toggle_role":
		_, err = h.authSvc.ToggleRole(id)
	case "reset_password":
		newPass := generateRandomPassword()
		err = h.authSvc.ResetPassword(id, newPass)
		if err == nil {
			toastMsg = fmt.Sprintf("Password reset to: %s", newPass)
		}
	}

	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	c.Header("HX-Toast-Success", toastMsg)
	h.renderUsersTableBody(c)
}

func (h *FrontendHandler) htmxDeleteUser(c *gin.Context) {
	role, _ := c.Get(ContextRoleKey)
	if role != "admin" {
		c.Status(http.StatusForbidden)
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	if err := h.authSvc.DeleteUser(id); err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	c.Header("HX-Toast-Success", "User deleted successfully")
	h.renderUsersTableBody(c)
}

func (h *FrontendHandler) renderUsersTableBody(c *gin.Context) {
	users, err := h.authSvc.ListUsers()
	if err != nil {
		users = []*entity.User{}
	}

	html := ""
	for _, u := range users {
		roleBadge := `<span class="status-badge status-success">user</span>`
		if u.Role == "admin" {
			roleBadge = `<span class="status-badge status-info">admin</span>`
		}
		totpIcon := `<i class="df-icon-cancel text-gray-500"></i>`
		if u.TOTPEnabled {
			totpIcon = `<i class="df-icon-check-circle text-success"></i>`
		}
		msIcon := `<i class="df-icon-cancel text-gray-500"></i>`
		if u.MSLinked {
			msIcon = `<i class="df-icon-check-circle text-success"></i>`
		}
		created := ""
		if u.CreatedAt != nil {
			created = u.CreatedAt.Format("Jan 02, 2006")
		}

		html += fmt.Sprintf(`<tr>
			<td class="ps-4">%s</td>
			<td>%s</td>
			<td>%s</td>
			<td>%s</td>
			<td>%s</td>
			<td>%s</td>
			<td class="pe-4"><div class="dropdown"><button class="btn btn-sm btn-link" data-bs-toggle="dropdown"><i class="df-icon-options-2"></i></button><ul class="dropdown-menu"><li><button class="dropdown-item" hx-patch="/app/htmx/users/%s" hx-vals='{"action":"toggle_role"}' hx-target="#users-table-body" hx-confirm="Change role for %s?"><i class="df-icon-shield-person me-1"></i> Toggle Role</button></li><li><button class="dropdown-item" hx-patch="/app/htmx/users/%s" hx-vals='{"action":"reset_password"}' hx-target="#users-table-body" hx-confirm="Reset password for %s?"><i class="df-icon-password me-1"></i> Reset Password</button></li><li><hr class="dropdown-divider"></li><li><button class="dropdown-item text-danger" hx-delete="/app/htmx/users/%s" hx-target="#users-table-body" hx-confirm="Delete user %s?"><i class="df-icon-delete me-1"></i> Delete</button></li></ul></div></td>
		</tr>`, htmlpkg.EscapeString(u.Username), htmlpkg.EscapeString(u.Email), roleBadge, totpIcon, msIcon, created, u.ID, htmlpkg.EscapeString(u.Username), u.ID, htmlpkg.EscapeString(u.Username), u.ID, htmlpkg.EscapeString(u.Username))
	}

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, html)
}

func buildQRRequest(c *gin.Context) *request.QRCode {
	width := 256
	if w := c.PostForm("qr_width"); w != "" {
		if parsed, err := strconv.Atoi(w); err == nil && parsed > 0 {
			width = parsed
		}
	}

	bgColor := c.DefaultPostForm("qr_bg_color", "#ffffff")
	if c.PostForm("qr_bg_transparent") == "on" {
		bgColor = "transparent"
	}

	if bgHex := c.PostForm("qr_bg_hex"); bgHex != "" && bgHex != bgColor {
		bgColor = bgHex
	}

	qr := &request.QRCode{
		Create:       true,
		Width:        width,
		Shape:        c.DefaultPostForm("qr_shape", "rect"),
		OutputFormat: c.DefaultPostForm("qr_format", "svg"),
		FgColor:      c.DefaultPostForm("qr_fg_color", "#000000"),
		BgColor:      bgColor,
	}

	// Handle logo upload
	if c.PostForm("qr_no_logo") == "on" {
		qr.LogoImage = "none" // Sentinel: skip default logo
	} else {
		file, header, err := c.Request.FormFile("qr_logo")
		if err == nil && header.Size > 0 {
			defer file.Close()
			if header.Size > maxLogoBytes {
				return qr
			}
			logoBytes, readErr := io.ReadAll(file)
			if readErr == nil && len(logoBytes) > 0 {
				qr.LogoImage = base64.StdEncoding.EncodeToString(logoBytes)
			}
		}
	}

	return qr
}

func renderBreakdownTable(c *gin.Context, items []service.BreakdownItem, columnName string) {
	if len(items) == 0 {
		c.String(http.StatusOK, fmt.Sprintf(`<p class="text-gray-500 text-center">No %s data available</p>`, columnName))
		return
	}

	html := fmt.Sprintf(`<table class="table table-sm"><thead><tr><th>%s</th><th class="text-end">Clicks</th></tr></thead><tbody>`, columnName)
	for _, item := range items {
		html += fmt.Sprintf(`<tr><td>%s</td><td class="text-end fw-semibold">%d</td></tr>`, htmlpkg.EscapeString(item.Name), item.Count)
	}
	html += `</tbody></table>`

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, html)
}

func generateRandomPassword() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
