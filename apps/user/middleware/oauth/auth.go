package oauth

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/iReflect/reflect-app/apps/user/models"
	"github.com/iReflect/reflect-app/apps/user/services"
	"github.com/iReflect/reflect-app/config"
)

// CookieAuthenticationMiddleware ...
func CookieAuthenticationMiddleware(service services.AuthenticationService) gin.HandlerFunc {
	return func(c *gin.Context) {
		status, err := service.AuthenticateSession(c)
		if err != nil {
			c.AbortWithStatus(status)
			return
		}

	}
}

// AdminCookieAuthenticationMiddleware ...
func AdminCookieAuthenticationMiddleware(service services.AuthenticationService) gin.HandlerFunc {
	return func(c *gin.Context) {
		_, err := service.AuthenticateSession(c)
		if err != nil {
			c.Redirect(http.StatusFound, config.GetConfig().Server.LoginURL)
			return
		}

		if user, exist := c.Get("user"); !exist || !user.(models.User).IsAdmin {
			c.Redirect(http.StatusFound, config.GetConfig().Server.LoginURL)
			return
		}
	}
}
