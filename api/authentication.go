package api

import (
	"net/http"

	"github.com/doutorfinancas/pun-sho/api/response"
	"github.com/gin-gonic/gin"
)

type AuthenticationMiddleware struct {
	token string
}

func NewAuthenticationMiddleware(token string) *AuthenticationMiddleware {
	return &AuthenticationMiddleware{
		token: token,
	}
}

func (a *AuthenticationMiddleware) Authenticated(c *gin.Context) {
	token := c.GetHeader("token")

	if token != a.token {
		c.AbortWithStatusJSON(http.StatusUnauthorized, response.NewFailure("unauthorized"))
		return
	}

	c.Next()
}
