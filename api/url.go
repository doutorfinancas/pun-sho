package api

import (
	"net/http"

	"github.com/doutorfinancas/pun-sho/convert"
	"github.com/gin-gonic/gin"
)

type urlHandler struct {
	unknownPage string
}

func NewURLHandler(unknownPage string) HTTPHandler {
	return &urlHandler{
		unknownPage: unknownPage,
	}
}

func (h *urlHandler) Routes(rg *gin.RouterGroup) {
	rg.GET("/:slug", h.RedirectLinkIfExists)
	rg.GET("/:slug/", h.RedirectLinkIfExists)
}

func (h *urlHandler) Group() *string {
	return convert.ToStringNil("s")
}

func (h *urlHandler) RedirectLinkIfExists(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		c.Redirect(http.StatusMovedPermanently, h.unknownPage)
		return
	}

	// @TODO Check if link exists
	// @TODO if it does, redirect

	c.Redirect(http.StatusMovedPermanently, h.unknownPage)
}
