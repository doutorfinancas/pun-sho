package api

const ResponseOk = "ok"
const ResponseError = "error"

type BaseResponse struct {
	Status string `json:"status"`
}

type ErrorResponse struct {
	BaseResponse
	Message []string `json:"message,omitempty"`
}

func NewErrorResponse(message string) *ErrorResponse {
	return &ErrorResponse{BaseResponse{Status: ResponseError}, []string{message}}
}

func NewOkResponse() *BaseResponse {
	return &BaseResponse{Status: ResponseOk}
}
