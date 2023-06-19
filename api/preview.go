package api

import (
	"net/http"

	"github.com/doutorfinancas/pun-sho/api/request"
	"github.com/doutorfinancas/pun-sho/api/response"
	"github.com/doutorfinancas/pun-sho/service"
	"github.com/doutorfinancas/pun-sho/str"
	"github.com/gin-gonic/gin"
)

type previewHandler struct {
	qrSvc *service.QRCodeService
}

func NewPreviewHandler(qrSvc *service.QRCodeService) HTTPHandler {
	return &previewHandler{
		qrSvc: qrSvc,
	}
}

func (h *previewHandler) Routes(rg *gin.RouterGroup) {
	rg.POST("", h.CreateLink)
	rg.POST("/", h.CreateLink)
}

func (h *previewHandler) Group() *string {
	return str.ToStringNil("preview")
}

func (h *previewHandler) CreateLink(c *gin.Context) {
	m := &request.GeneratePreview{}
	err := c.BindJSON(m)

	if err != nil {
		c.JSON(
			http.StatusBadRequest,
			response.NewFailure("invalid payload"),
		)
		return
	}
	s, err := h.qrSvc.Generate(m.QRCode, m.Link)
	if err != nil {
		c.JSON(
			http.StatusBadRequest,
			response.NewFailure("kaput, no save"),
		)
		return
	}

	a := response.NewGeneratePreviewResponse(s, nil)

	c.JSON(http.StatusCreated, a)
}
