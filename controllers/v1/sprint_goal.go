package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
	retrospectiveServices "github.com/iReflect/reflect-app/apps/retrospective/services"
)

// SprintGoalController ...
type SprintGoalController struct {
	SprintService     retrospectiveServices.SprintService
	PermissionService retrospectiveServices.PermissionService
	TrailService      retrospectiveServices.TrailService
}

// Routes for Sprints
func (ctrl SprintGoalController) Routes(r *gin.RouterGroup) {
	r.POST("/", ctrl.Add)
	r.GET("/", ctrl.List)
	r.PUT("/:goalID/", ctrl.Update)
	r.POST("/:goalID/resolve", ctrl.Resolve)
	r.DELETE("/:goalID/resolve", ctrl.UnResolve)
}

// Add Goal to sprint's retrospective
func (ctrl SprintGoalController) Add(c *gin.Context) {
	userID, _ := c.Get("userID")
	sprintID := c.Param("sprintID")
	retroID := c.Param("retroID")

	if !ctrl.PermissionService.UserCanEditSprint(retroID, sprintID, userID.(uint)) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{})
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}

// List goals associated to sprint
func (ctrl SprintGoalController) List(c *gin.Context) {
	userID, _ := c.Get("userID")
	sprintID := c.Param("sprintID")
	retroID := c.Param("retroID")
	if !ctrl.PermissionService.UserCanAccessSprint(retroID, sprintID, userID.(uint)) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{})
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}

// Update goal associated to a sprint
func (ctrl SprintGoalController) Update(c *gin.Context) {
	userID, _ := c.Get("userID")
	sprintID := c.Param("sprintID")
	retroID := c.Param("retroID")
	//goalID := c.Param("goalID")

	if !ctrl.PermissionService.UserCanEditSprint(retroID, sprintID, userID.(uint)) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{})
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}

// Resolve goal associated to a sprint
func (ctrl SprintGoalController) Resolve(c *gin.Context) {
	userID, _ := c.Get("userID")
	sprintID := c.Param("sprintID")
	retroID := c.Param("retroID")
	//goalID := c.Param("goalID")

	if !ctrl.PermissionService.UserCanEditSprint(retroID, sprintID, userID.(uint)) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{})
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}

// UnResolve a goal associated to a sprint
func (ctrl SprintGoalController) UnResolve(c *gin.Context) {
	userID, _ := c.Get("userID")
	sprintID := c.Param("sprintID")
	retroID := c.Param("retroID")
	//goalID := c.Param("goalID")

	if !ctrl.PermissionService.UserCanEditSprint(retroID, sprintID, userID.(uint)) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{})
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}