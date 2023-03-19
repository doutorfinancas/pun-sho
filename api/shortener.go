package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/doutorfinancas/pun-sho/api/request"
	"github.com/doutorfinancas/pun-sho/service"
	"github.com/doutorfinancas/pun-sho/str"
)

type shortenerHandler struct {
	service *service.ShortyService
}

func NewShortenerHandler(svc *service.ShortyService) HTTPHandler {
	return &shortenerHandler{service: svc}
}

func (h *shortenerHandler) Routes(rg *gin.RouterGroup) {
	rg.GET("/:id", h.GetLinkInformation)
	rg.GET("", h.ListLinks)
	rg.POST("", h.CreateLink)
	rg.DELETE("", h.RemoveLink)
}

func (h *shortenerHandler) Group() *string {
	return str.ToStringNil("short")
}

func (h *shortenerHandler) GetLinkInformation(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(
			http.StatusBadRequest,
			NewErrorResponse("no id provided"),
		)
	}
	parsed := uuid.MustParse(id)
	shorty, err := h.service.FindShortyByID(parsed)
	if err != nil {
		c.JSON(http.StatusNotFound, "shorty not found")
		return
	}

	c.JSON(http.StatusOK, shorty)
}

func (h *shortenerHandler) ListLinks(c *gin.Context) {
	limitStr := c.Query("limit")
	offsetStr := c.Query("offset")

	limit, offset, message, err := validateLimitAndOffset(limitStr, offsetStr)
	if err != nil {
		c.JSON(
			http.StatusBadRequest,
			NewErrorResponse(message),
		)
		return
	}

	links, err := h.service.List(limit, offset)
	if err != nil {
		c.JSON(
			http.StatusBadRequest,
			NewErrorResponse("kaput, no links for you"),
		)
		return
	}

	c.JSON(http.StatusOK, links)
}

func (h *shortenerHandler) CreateLink(c *gin.Context) {
	m := &request.CreateShorty{}
	err := c.BindJSON(m)

	if err != nil {
		c.JSON(
			http.StatusBadRequest,
			NewErrorResponse("invalid payload"),
		)
		return
	}
	s, err := h.service.Create(m)
	if err != nil {
		c.JSON(
			http.StatusBadRequest,
			NewErrorResponse("kaput, no save"),
		)
		return
	}

	c.JSON(http.StatusCreated, s)
}

func (h *shortenerHandler) RemoveLink(c *gin.Context) {
	// @TODO Implement me!
	c.JSON(
		http.StatusNotImplemented,
		NewErrorResponse("nope, not yet. Try again later, boss"),
	)
}
