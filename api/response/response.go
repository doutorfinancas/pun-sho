package response

const Ok = "ok"
const Error = "error"

type BaseResponse struct {
	Status string `json:"status"`
}

type FailureResponse struct {
	BaseResponse
	Message []string `json:"message,omitempty"`
}

func NewFailure(message string) *FailureResponse {
	return &FailureResponse{BaseResponse{Status: Error}, []string{message}}
}

func NewOk() *BaseResponse {
	return &BaseResponse{Status: Ok}
}
