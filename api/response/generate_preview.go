package response

type GeneratePreviewResponse struct {
	BaseResponse
	QrCode  string    `json:"qr_code"`
	Message *[]string `json:"message,omitempty"`
}

// NewGeneratePreviewResponse creates a base preview response
func NewGeneratePreviewResponse(qrCode string, message *[]string) *GeneratePreviewResponse {
	return &GeneratePreviewResponse{
		BaseResponse{Status: Ok},
		qrCode,
		message,
	}
}
