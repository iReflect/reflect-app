package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

//FeedbackController ...
type UserController struct {
}

// Routes for Feedback
func (ctrl UserController) Routes(r *gin.RouterGroup) {
	r.GET("/current/", ctrl.Current)
}

// Get feedback
func (ctrl UserController) Current(c *gin.Context) {
	user, ok := c.Get("user")
	if !ok {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}
