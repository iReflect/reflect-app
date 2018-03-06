package controllers

import (
	"github.com/gin-gonic/gin"
	userServices "github.com/iReflect/reflect-app/apps/user/services"
	"net/http"
)

//UserAuthController ...
type UserAuthController struct {
	AuthService userServices.AuthenticationService
}

//Add Routes
func (ctrl UserAuthController) Routes(r *gin.RouterGroup) {
	r.GET("/login/", ctrl.Login)
	r.POST("/auth/", ctrl.Auth)
	r.POST("/logout/", ctrl.Logout)
}

// Login ...
func (ctrl UserAuthController) Login(c *gin.Context) {
	oauthRequest := ctrl.AuthService.Login(c)
	c.JSON(http.StatusOK, oauthRequest)
}

// Auth ...
func (ctrl UserAuthController) Auth(c *gin.Context) {
	user, status, err := ctrl.AuthService.Authorize(c)
	if err != nil {
		c.AbortWithStatusJSON(status, gin.H{"error": err})
		return
	}
	c.JSON(status, user)
}

// Logout ...
func (ctrl UserAuthController) Logout(c *gin.Context) {
	status := ctrl.AuthService.Logout(c)
	if status == http.StatusOK {
		c.AbortWithStatusJSON(http.StatusOK, gin.H{"message": "success"})
		return
	}
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
}
