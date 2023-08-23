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
	rg.POST("", h.CreatePreview)
	rg.POST("/", h.CreatePreview)
}

func (h *previewHandler) Group() *string {
	return str.ToStringNil("preview")
}

// CreatePreview godoc
// @Tags Preview
// @Summary Creates a QR Code preview for a given url
// @Schemes
// @Description Creates a QR Code preview for a given url
// @Param token header string false "Authorization token"
// @Param request body request.GeneratePreview true "Request"
// @Produce json
// @Success 201 {object} response.GeneratePreviewResponse "response"
// @Failure 400 {object} response.FailureResponse "error"
// @Router /preview [post]
func (h *previewHandler) CreatePreview(c *gin.Context) {
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
			response.NewFailure("failed to generate preview"),
		)
		return
	}

	a := response.NewGeneratePreviewResponse(s, nil)

	c.JSON(http.StatusCreated, a)
}
