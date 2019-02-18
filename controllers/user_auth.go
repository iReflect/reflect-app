package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	userServices "github.com/iReflect/reflect-app/apps/user/services"
)

//UserAuthController ...
type UserAuthController struct {
	AuthService userServices.AuthenticationService
}

// Routes for UserAuthController
func (ctrl UserAuthController) Routes(r *gin.RouterGroup) {
	r.GET("/login/", ctrl.Login)
	r.POST("/login/", ctrl.BasicLogin)
	r.POST("/identify/", ctrl.Identify)
	// TODO make auth get and receive request directly from google
	r.POST("/auth/", ctrl.Auth)
	r.POST("/logout/", ctrl.Logout)
}

// ToDo: handle errors like in retrospectives/sprints controllers

// Login ...
func (ctrl UserAuthController) Login(c *gin.Context) {
	oauthRequest := ctrl.AuthService.Login(c)
	c.JSON(http.StatusOK, oauthRequest)
}

// BasicLogin ...
func (ctrl UserAuthController) BasicLogin(c *gin.Context) {
	user, status, err := ctrl.AuthService.BasicLogin(c)
	if err != nil {
		c.AbortWithStatusJSON(status, gin.H{"error": err.Error()})
		return
	}
	c.JSON(status, user)
}

// Identify ...
func (ctrl UserAuthController) Identify(c *gin.Context) {
	status, err := ctrl.AuthService.Identify(c)
	if err != nil {
		c.AbortWithStatusJSON(status, gin.H{"error": err.Error()})
		return
	}
	c.JSON(status, nil)
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
	if status == http.StatusOK {
		c.AbortWithStatusJSON(http.StatusOK, gin.H{"message": "success"})
		return
	}
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
}
