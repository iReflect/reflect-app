package oauth

import (
	"github.com/gin-gonic/gin"
	"github.com/iReflect/reflect-app/apps/user/models"
	"github.com/iReflect/reflect-app/apps/user/services"
	"github.com/iReflect/reflect-app/config"
	"net/http"
)

// CookieAuthenticationMiddleware ...
func CookieAuthenticationMiddleware(service services.AuthenticationService) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !service.AuthenticateSession(c) {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

	}
}

// AdminCookieAuthenticationMiddleware ...
func AdminCookieAuthenticationMiddleware(service services.AuthenticationService) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !service.AuthenticateSession(c) {
			c.Redirect(http.StatusFound, config.GetConfig().Server.LoginURL)
			return
		}

		if user, exist := c.Get("user"); !exist || !user.(models.User).IsAdmin {
			c.Redirect(http.StatusFound, config.GetConfig().Server.LoginURL)
			return
		}
	}
}
