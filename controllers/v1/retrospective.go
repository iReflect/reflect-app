package v1

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	retrospectiveService "github.com/iReflect/reflect-app/apps/retrospective/services"
)

// RetroSpectiveController ...
type RetrospectiveController struct {
	RetrospectiveService retrospectiveService.RetrospectiveService
	PermissionService    retrospectiveService.PermissionService
}

// Routes for RetroSpective
func (ctrl RetrospectiveController) Routes(r *gin.RouterGroup) {
	r.GET("/", ctrl.List)
	r.GET("/:retroID", ctrl.Get)
}

// List RetroSpectives
func (ctrl RetrospectiveController) List(c *gin.Context) {
	userID, _ := c.Get("userID")
	perPage, err := strconv.Atoi(c.DefaultQuery("perPage", ""))
	if err != nil {
		perPage = -1
	}
	page, err := strconv.Atoi(c.DefaultQuery("page", ""))
	if err != nil {
		page = 1
	}
	response, err := ctrl.RetrospectiveService.List(userID.(uint), perPage, page)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "Retrospectives not found", "error": err})
		return
	}
	c.JSON(http.StatusOK, response)
}

// Get RetroSpective by id
func (ctrl RetrospectiveController) Get(c *gin.Context) {
	retrospectiveID := c.Param("retroID")
	userID, _ := c.Get("userID")

	if !ctrl.PermissionService.UserCanAccessRetro(retrospectiveID, userID.(uint)) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{})
		return
	}
	response, err := ctrl.RetrospectiveService.Get(retrospectiveID)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "Retrospective not found", "error": err})
		return
	}
	c.JSON(http.StatusOK, response)
}
