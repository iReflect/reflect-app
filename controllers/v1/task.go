package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
	retroServices "github.com/iReflect/reflect-app/apps/retrospective/services"
)

// TaskController ...
type TaskController struct {
	TaskService       retroServices.TaskService
	PermissionService retroServices.PermissionService
}

// Routes for Tasks
func (ctrl TaskController) Routes(r *gin.RouterGroup) {
	r.GET("/", ctrl.List)
	r.GET("/:id", ctrl.Get)
	r.GET("/:id/members", ctrl.GetMembers)
}

// List ...
func (ctrl TaskController) List(c *gin.Context) {
	retroID := c.Param("retroID")
	sprintID := c.Param("sprintID")
	userID, _ := c.Get("userID")

	if !ctrl.PermissionService.UserCanAccessSprint(retroID, userID.(uint)) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{})
		return
	}

	tasks, err := ctrl.TaskService.List(retroID, sprintID)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "Could not get tasks", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tasks)
}

// Get ...
func (ctrl TaskController) Get(c *gin.Context) {
	id := c.Param("id")
	retroID := c.Param("retroID")
	sprintID := c.Param("sprintID")
	userID, _ := c.Get("userID")

	if !ctrl.PermissionService.UserCanAccessTask(retroID, userID.(uint)) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{})
		return
	}

	task, err := ctrl.TaskService.Get(id, retroID, sprintID)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "Could not get task", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, task)
}

// GetMembers ...
func (ctrl TaskController) GetMembers(c *gin.Context) {
	id := c.Param("id")
	retroID := c.Param("retroID")
	sprintID := c.Param("sprintID")
	userID, _ := c.Get("userID")

	if !ctrl.PermissionService.UserCanAccessTask(id, userID.(uint)) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{})
		return
	}

	members, err := ctrl.TaskService.GetMembers(id, retroID, sprintID)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "Could not get members", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, members)
}
