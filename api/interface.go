package api

import (
	"github.com/gin-gonic/gin"

	"github.com/doutorfinancas/pun-sho/str"
)

type HTTPHandler interface {
	Routes(rg *gin.RouterGroup)
	Group() *string
}

type Server interface {
	PushHandlerWithGroup(h HTTPHandler, rg *gin.RouterGroup)
}

type BaseGinServer struct{}

func (*BaseGinServer) PushHandlerWithGroup(h HTTPHandler, rg *gin.RouterGroup) {
	if gs := str.ToString(h.Group()); gs != "" {
		h.Routes(rg.Group(gs))

		return
	}

	h.Routes(rg)
}
