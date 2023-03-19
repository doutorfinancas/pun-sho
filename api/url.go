package api

import (
	"net/http"

	"github.com/doutorfinancas/pun-sho/service"
	"github.com/gin-gonic/gin"

	"github.com/doutorfinancas/pun-sho/str"
)

type urlHandler struct {
	unknownPage string
	service     *service.ShortyService
}

func NewURLHandler(unknownPage string, svc *service.ShortyService) HTTPHandler {
	return &urlHandler{
		unknownPage: unknownPage,
		service:     svc,
	}
}

func (h *urlHandler) Routes(rg *gin.RouterGroup) {
	rg.GET("/:slug", h.RedirectLinkIfExists)
	rg.GET("/:slug/", h.RedirectLinkIfExists)
}

func (h *urlHandler) Group() *string {
	return str.ToStringNil("s")
}

func (h *urlHandler) RedirectLinkIfExists(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		c.Redirect(http.StatusMovedPermanently, h.unknownPage)
		return
	}

	sho, err := h.service.CreateVisit(slug, c.Request.UserAgent())
	if err != nil {
		c.Redirect(http.StatusMovedPermanently, h.unknownPage)
	}

	c.Redirect(http.StatusMovedPermanently, sho.Link)
}
