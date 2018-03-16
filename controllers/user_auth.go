package controllers

import (
	"github.com/gin-gonic/gin"
	"net/http"

	userServices "github.com/iReflect/reflect-app/apps/user/services"
	"github.com/iReflect/reflect-app/constants"
)

//UserAuthController ...
type UserAuthController struct {
	AuthService userServices.AuthenticationService
}

//Add Routes
func (ctrl UserAuthController) Routes(r *gin.RouterGroup) {
	r.GET("/login/", ctrl.Login)
	// TODO make auth get and receive request directly from google
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
		c.AbortWithStatusJSON(status, gin.H{"error": err.Error()})
		return
	}
	c.JSON(status, user)
}

// Logout ...
func (ctrl UserAuthController) Logout(c *gin.Context) {
	status := ctrl.AuthService.Logout(c)
	if status == http.StatusNoContent {
		c.AbortWithStatusJSON(status, gin.H{})
		return
	}
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": constants.UnAuthorizedUser})
}
