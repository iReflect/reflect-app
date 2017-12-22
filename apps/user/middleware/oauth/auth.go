package oauth

import (
	"github.com/gin-gonic/gin"
	"github.com/iReflect/reflect-app/apps/user/services"
	"net/http"
)

func CookieAuthenticationMiddleWare(service services.AuthenticationService) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !service.AuthenticateSession(c) {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

	}
}
