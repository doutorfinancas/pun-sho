package api

import (
	"net/http"

	"github.com/doutorfinancas/pun-sho/api/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/doutorfinancas/pun-sho/api/request"
	"github.com/doutorfinancas/pun-sho/service"
	"github.com/doutorfinancas/pun-sho/str"
)

type shortenerHandler struct {
	shortySvc *service.ShortyService
}

func NewShortenerHandler(shortySvc *service.ShortyService) HTTPHandler {
	return &shortenerHandler{
		shortySvc: shortySvc,
	}
}

func (h *shortenerHandler) Routes(rg *gin.RouterGroup) {
	rg.GET("/:id", h.GetLinkInformation)
	rg.GET("", h.ListLinks)
	rg.POST("", h.CreateLink)
	rg.PATCH("/:id", h.EditLink)
	rg.DELETE("/:id", h.RemoveLink)
}

func (h *shortenerHandler) Group() *string {
	return str.ToStringNil("short")
}

// GetLinkInformation godoc
// @Tags Short
// @Summary get your shortlink information
// @Schemes
// @Description retrieves full information for the give shortlink
// @Param token header string false "Authorization token"
// @Param id path string true "ShortLink ID"
// @Param from query string true "accesses from date 'YYYY-mm-dd'"
// @Param until query string true "accesses until date 'YYYY-mm-dd'"
// @Success 200 {object} entity.Shorty "response"
// @Failure 400 {object} response.FailureResponse "error"
// @Failure 404 {object} response.FailureResponse "not found"
// @Router /short/{id} [get]
func (h *shortenerHandler) GetLinkInformation(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(
			http.StatusBadRequest,
			response.NewFailure("no id provided"),
		)
	}
	parsed := uuid.MustParse(id)
	from := c.Query("from")
	until := c.Query("until")
	shorty, err := h.shortySvc.FindShortyByID(parsed, from, until)
	if err != nil {
		c.JSON(http.StatusNotFound, "shorty not found")
		return
	}

	c.JSON(http.StatusOK, shorty)
}

// ListLinks godoc
// @Tags Short
// @Summary Lists your shorlinks
// @Schemes
// @Description Lists all the shortlinks available
// @Param token header string false "Authorization token"
// @Produce json
// @Success 200 {object} []entity.Shorty "response"
// @Failure 400 {object} response.FailureResponse "error"
// @Router /short [get]
func (h *shortenerHandler) ListLinks(c *gin.Context) {
	limitStr := c.Query("limit")
	offsetStr := c.Query("offset")

	limit, offset, message, err := validateLimitAndOffset(limitStr, offsetStr)
	if err != nil {
		c.JSON(
			http.StatusBadRequest,
			response.NewFailure(message),
		)
		return
	}

	links, err := h.shortySvc.List(limit, offset)
	if err != nil {
		c.JSON(
			http.StatusBadRequest,
			response.NewFailure("kaput, no links for you"),
		)
		return
	}

	c.JSON(http.StatusOK, links)
}

// CreateLink godoc
// @Tags Short
// @Summary Creates a shortlink for a given url
// @Schemes
// @Description Creates a shortlink for a given url, optionally setting a ttl and a redirection limit
// @Param token header string false "Authorization token"
// @Param request body request.CreateShorty true "Request"
// @Produce json
// @Success 201 {object} entity.Shorty "response"
// @Failure 400 {object} response.FailureResponse "error"
// @Router /short [post]
func (h *shortenerHandler) CreateLink(c *gin.Context) {
	m := &request.CreateShorty{}
	err := c.BindJSON(m)
	if err != nil {
		c.JSON(
			http.StatusBadRequest,
			response.NewFailure("invalid payload"),
		)
		return
	}

	s, err := h.shortySvc.Create(m)
	if err != nil {
		c.JSON(
			http.StatusBadRequest,
			response.NewFailure("kaput, no save"),
		)
		return
	}

	c.JSON(http.StatusCreated, s)
}

// EditLink godoc
// @Tags Short
// @Summary Edits a shortlink
// @Schemes
// @Description Edits a shortlink, allowing to set TTL, cancel the link or change the redirection limit or associated link
// @Param token header string false "Authorization token"
// @Param id path string true "ShortLink ID"
// @Param request body request.UpdateShorty true "Request"
// @Produce json
// @Success 200 {object} entity.Shorty "response"
// @Failure 400 {object} response.FailureResponse "error"
// @Failure 404 {object} response.FailureResponse "not found"
// @Router /short/{id} [patch]
func (h *shortenerHandler) EditLink(c *gin.Context) {
	id := c.Param("id")
	m := &request.UpdateShorty{}
	err := c.BindJSON(m)
	if err != nil || id == "" {
		c.JSON(
			http.StatusBadRequest,
			response.NewFailure("invalid payload"),
		)
		return
	}

	shorty, err := h.shortySvc.FindShortyByID(uuid.MustParse(id), "", "")
	if err != nil {
		c.JSON(http.StatusNotFound, "shorty not found")
		return
	}

	updatedShorty, err := h.shortySvc.Update(m, shorty)
	if err != nil {
		c.JSON(
			http.StatusBadRequest,
			response.NewFailure("kaput, no save"),
		)
		return
	}

	c.JSON(http.StatusOK, updatedShorty)
}

// RemoveLink godoc
// @Tags Short
// @Summary Deletes a shortlink
// @Schemes
// @Description Deletes a shortlink
// @Param token header string false "Authorization token"
// @Param id path string true "ShortLink ID"
// @Param request body request.UpdateShorty true "Request"
// @Produce json
// @Success 204 string false ""
// @Failure 400 {object} response.FailureResponse "error"
// @Failure 404 {object} response.FailureResponse "not found"
// @Router /short/{id} [delete]
func (h *shortenerHandler) RemoveLink(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(
			http.StatusBadRequest,
			response.NewFailure("no id provided"),
		)
	}
	parsed := uuid.MustParse(id)
	err := h.shortySvc.DeleteShortyByUUID(parsed)
	if err != nil {
		c.JSON(
			http.StatusBadRequest,
			response.NewFailure("kaput, no delete"),
		)
	}
	c.JSON(http.StatusNoContent, nil)
}
