package v1

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	retroSpectiveService "github.com/iReflect/reflect-app/apps/retrospective/services"
)

// RetroSpectiveController ...
type RetroSpectiveController struct {
	RetroSpectiveService retroSpectiveService.RetroSpectiveService
}

// Routes for RetroSpective
func (ctrl RetroSpectiveController) Routes(r *gin.RouterGroup) {
	r.GET("/", ctrl.List)
}

// List RetroSpectives
func (ctrl RetroSpectiveController) List(c *gin.Context) {
	userID, _ := c.Get("userID")
	perPage, err := strconv.Atoi(c.DefaultQuery("perPage", ""))
	if err != nil {
		perPage = -1
	}
	page, err := strconv.Atoi(c.DefaultQuery("page", ""))
	if err != nil {
		page = 1
	}
	response, err := ctrl.RetroSpectiveService.List(userID.(uint), perPage, page)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "Retrospectives not found", "error": err})
		return
	}
	c.JSON(http.StatusOK, response)
}
