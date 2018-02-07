package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

//UserController ...
type UserController struct {
}

// Routes for User
func (ctrl UserController) Routes(r *gin.RouterGroup) {
	r.GET("/current/", ctrl.Current)
}

// Get current user
func (ctrl UserController) Current(c *gin.Context) {
	user, ok := c.Get("user")
	if !ok {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}
