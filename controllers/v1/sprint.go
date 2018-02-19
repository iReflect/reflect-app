package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gocraft/work"
	retrospectiveServices "github.com/iReflect/reflect-app/apps/retrospective/services"
	"github.com/iReflect/reflect-app/workers"
)

// SprintController ...
type SprintController struct {
	SprintService     retrospectiveServices.SprintService
	PermissionService retrospectiveServices.PermissionService
}

// Routes for Sprints
func (ctrl SprintController) Routes(r *gin.RouterGroup) {
	r.GET("/", ctrl.GetSprints)
	r.DELETE("/:sprintID", ctrl.Delete)
	r.GET("/:sprintID", ctrl.Get)
	r.POST("/:sprintID/activate", ctrl.ActivateSprint)
	r.POST("/:sprintID/process", ctrl.Process)
}

// GetSprints ...
func (ctrl SprintController) GetSprints(c *gin.Context) {
	userID, _ := c.Get("userID")
	retroID := c.Param("retroID")
	if !ctrl.PermissionService.UserCanAccessRetro(retroID, userID.(uint)) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{})
		return
	}

	sprints, err := ctrl.SprintService.GetSprintsList(retroID, userID.(uint))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "Failed to get sprints", "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, sprints)
}

// Delete Sprint
func (ctrl SprintController) Delete(c *gin.Context) {
	userID, _ := c.Get("userID")
	sprintID := c.Param("sprintID")
	if !ctrl.PermissionService.UserCanEditSprint(sprintID, userID.(uint)) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{})
		return
	}

	if err := ctrl.SprintService.DeleteSprint(sprintID); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "Sprint couldn't be deleted", "error": err.Error()})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

// ActivateSprint activates the given sprint
func (ctrl SprintController) ActivateSprint(c *gin.Context) {
	userID, _ := c.Get("userID")
	sprintID := c.Param("sprintID")
	if !ctrl.PermissionService.UserCanEditSprint(sprintID, userID.(uint)) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{})
		return
	}

	if err := ctrl.SprintService.ActivateSprint(sprintID); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "Sprint couldn't be activated", "error": err.Error()})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

// Get Sprint Data
func (ctrl SprintController) Get(c *gin.Context) {
	userID, _ := c.Get("userID")
	sprintID := c.Param("sprintID")
	if !ctrl.PermissionService.UserCanAccessSprint(sprintID, userID.(uint)) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{})
		return
	}

	sprint, err := ctrl.SprintService.Get(sprintID, userID.(uint))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "Failed to get sprint data", "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, sprint)
}

// Process Sprint
func (ctrl SprintController) Process(c *gin.Context) {
	userID, _ := c.Get("userID")
	sprintID := c.Param("sprintID")
	if !ctrl.PermissionService.UserCanAccessSprint(sprintID, userID.(uint)) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{})
		return
	}

	workers.Enqueuer.EnqueueUnique("sync_sprint_data", work.Q{"sprintID": sprintID})

	c.JSON(http.StatusNoContent, nil)
}
