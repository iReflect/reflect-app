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
	r.GET("/login", ctrl.Login)
	r.GET("/auth", ctrl.Auth)
	r.GET("/logout", ctrl.Logout)
}

// Login ...
func (ctrl UserAuthController) Login(c *gin.Context) {
	state := ctrl.AuthService.Login(c)
	c.JSON(http.StatusOK, state) // TODO Replace with oauth request
}

// Auth ...
func (ctrl UserAuthController) Auth(c *gin.Context) {
	user, err := ctrl.AuthService.Authorize(c)
	if err != nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	c.JSON(http.StatusOK, user)
}

// Logout ...
func (ctrl UserAuthController) Logout(c *gin.Context) {
	c.AbortWithStatus(ctrl.AuthService.Logout(c))
}
