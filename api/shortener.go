package api

import (
	"net/http"

	"github.com/doutorfinancas/pun-sho/convert"
	"github.com/gin-gonic/gin"
)

type shortenerHandler struct {
}

func NewShortenerHandler() HTTPHandler {
	return &shortenerHandler{}
}

func (h *shortenerHandler) Routes(rg *gin.RouterGroup) {
	rg.GET("/:id", h.GetLinkInformation)
	rg.GET("", h.ListLinks)
	rg.POST("", h.CreateLink)
	rg.DELETE("", h.RemoveLink)
}

func (h *shortenerHandler) Group() *string {
	return convert.ToStringNil("short")
}

func (h *shortenerHandler) GetLinkInformation(c *gin.Context) {
	// @TODO Implement me!
	c.JSON(
		http.StatusNotImplemented,
		NewErrorResponse("nope, not yet. Try again later, boss"),
	)
}

func (h *shortenerHandler) ListLinks(c *gin.Context) {
	// @TODO Implement me!
	c.JSON(
		http.StatusNotImplemented,
		NewErrorResponse("nope, not yet. Try again later, boss"),
	)
}

func (h *shortenerHandler) CreateLink(c *gin.Context) {
	// @TODO Implement me!
	c.JSON(
		http.StatusNotImplemented,
		NewErrorResponse("nope, not yet. Try again later, boss"),
	)
}

func (h *shortenerHandler) RemoveLink(c *gin.Context) {
	// @TODO Implement me!
	c.JSON(
		http.StatusNotImplemented,
		NewErrorResponse("nope, not yet. Try again later, boss"),
	)
}
