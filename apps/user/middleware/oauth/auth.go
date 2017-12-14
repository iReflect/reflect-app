package oauth

import (
	"github.com/gin-gonic/gin"
	"github.com/iReflect/reflect-app/apps/user/services"
	"net/http"
)

func TokenAuthenticationMiddleWare(service services.AuthenticationService) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !service.Authenticate(c) {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

	}
}
