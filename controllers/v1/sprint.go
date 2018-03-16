package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gocraft/work"
	retroSerializers "github.com/iReflect/reflect-app/apps/retrospective/serializers"
	retrospectiveServices "github.com/iReflect/reflect-app/apps/retrospective/services"
	"github.com/iReflect/reflect-app/workers"
	"strconv"
)

// SprintController ...
type SprintController struct {
	SprintService     retrospectiveServices.SprintService
	PermissionService retrospectiveServices.PermissionService
	TrailService      retrospectiveServices.TrailService
}

// Routes for Sprints
func (ctrl SprintController) Routes(r *gin.RouterGroup) {
	r.GET("/", ctrl.GetSprints)
	r.POST("/", ctrl.CreateNewSprint)
	r.DELETE("/:sprintID/", ctrl.Delete)
	r.GET("/:sprintID/", ctrl.Get)
	r.PUT("/:sprintID/", ctrl.UpdateSprint)

	r.POST("/:sprintID/activate/", ctrl.ActivateSprint)
	r.POST("/:sprintID/freeze/", ctrl.FreezeSprint)
	r.POST("/:sprintID/process/", ctrl.Process)

	r.GET("/:sprintID/member-summary/", ctrl.GetSprintMemberSummary)
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
	retroID := c.Param("retroID")
	if !ctrl.PermissionService.UserCanEditSprint(retroID, sprintID, userID.(uint)) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{})
		return
	}

	if err := ctrl.SprintService.DeleteSprint(sprintID); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "Sprint couldn't be deleted", "error": err.Error()})
		return
	}

	ctrl.TrailService.Add("Deleted Sprint", "Sprint", sprintID, userID.(uint))

	c.JSON(http.StatusNoContent, nil)
}

// UpdateSprint ...
func (ctrl SprintController) UpdateSprint(c *gin.Context) {
	userID, _ := c.Get("userID")
	sprintID := c.Param("sprintID")
	retroID := c.Param("retroID")
	if !ctrl.PermissionService.UserCanEditSprint(retroID, sprintID, userID.(uint)) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{})
		return
	}
	var sprintData retroSerializers.UpdateSprintSerializer
	if err := c.BindJSON(&sprintData); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "Invalid request data", "error": err.Error()})
		return
	}

	response, err := ctrl.SprintService.UpdateSprint(sprintID, sprintData)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "Sprint couldn't be updated", "error": err.Error()})
		return
	}

	ctrl.TrailService.Add("Updated Sprint", "Sprint", sprintID, userID.(uint))

	c.JSON(http.StatusOK, response)
}

// ActivateSprint activates the given sprint
func (ctrl SprintController) ActivateSprint(c *gin.Context) {
	userID, _ := c.Get("userID")
	sprintID := c.Param("sprintID")
	retroID := c.Param("retroID")
	if !ctrl.PermissionService.UserCanEditSprint(retroID, sprintID, userID.(uint)) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{})
		return
	}

	if err := ctrl.SprintService.ActivateSprint(sprintID); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "Sprint couldn't be activated", "error": err.Error()})
		return
	}

	ctrl.TrailService.Add("Activated Sprint", "Sprint", sprintID, userID.(uint))

	c.JSON(http.StatusNoContent, nil)
}

// FreezeSprint freezes the given sprint
func (ctrl SprintController) FreezeSprint(c *gin.Context) {
	userID, _ := c.Get("userID")
	sprintID := c.Param("sprintID")
	retroID := c.Param("retroID")
	if !ctrl.PermissionService.UserCanEditSprint(retroID, sprintID, userID.(uint)) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{})
		return
	}

	if err := ctrl.SprintService.FreezeSprint(sprintID); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "Sprint couldn't be frozen", "error": err.Error()})
		return
	}

	ctrl.TrailService.Add("Froze Sprint", "Sprint", sprintID, userID.(uint))

	c.JSON(http.StatusNoContent, nil)
}

// Get Sprint Data
func (ctrl SprintController) Get(c *gin.Context) {
	userID, _ := c.Get("userID")
	sprintID := c.Param("sprintID")
	retroID := c.Param("retroID")
	if !ctrl.PermissionService.UserCanAccessSprint(retroID, sprintID, userID.(uint)) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{})
		return
	}

	sprint, err := ctrl.SprintService.Get(sprintID)
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
	retroID := c.Param("retroID")
	if !ctrl.PermissionService.UserCanAccessSprint(retroID, sprintID, userID.(uint)) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{})
		return
	}
	workers.Enqueuer.EnqueueUnique("sync_sprint_data", work.Q{"sprintID": sprintID})

	ctrl.TrailService.Add("Triggered Sprint Refresh", "Sprint", sprintID, userID.(uint))

	c.JSON(http.StatusNoContent, nil)
}

// GetSprintMemberSummary returns the sprint member summary list
func (ctrl SprintController) GetSprintMemberSummary(c *gin.Context) {
	userID, _ := c.Get("userID")
	sprintID := c.Param("sprintID")
	retroID := c.Param("retroID")
	if !ctrl.PermissionService.UserCanAccessSprint(retroID, sprintID, userID.(uint)) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{})
		return
	}
	response, err := ctrl.SprintService.GetSprintMembersSummary(sprintID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "Failed to get sprint member summary", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// CreateNewSprint creates a new draft sprint for the retro
func (ctrl SprintController) CreateNewSprint(c *gin.Context) {
	userID, _ := c.Get("userID")
	retroID := c.Param("retroID")
	if !ctrl.PermissionService.UserCanAccessRetro(retroID, userID.(uint)) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{})
		return
	}
	sprintData := retroSerializers.CreateSprintSerializer{CreatedByID: userID.(uint)}
	err := c.BindJSON(&sprintData)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "Invalid request data", "error": err.Error()})
		return
	}
	sprint, err := ctrl.SprintService.Create(retroID, sprintData)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "Failed to create the sprint", "error": err.Error()})
		return
	}

	ctrl.TrailService.Add("Created Sprint", "Sprint", strconv.Itoa(int(sprint.ID)), userID.(uint))

	c.JSON(http.StatusCreated, sprint)
}
