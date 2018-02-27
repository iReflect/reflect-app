package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
	retroSerializers "github.com/iReflect/reflect-app/apps/retrospective/serializers"
	retroServices "github.com/iReflect/reflect-app/apps/retrospective/services"
	"strconv"
)

// TaskController ...
type TaskController struct {
	TaskService       retroServices.TaskService
	PermissionService retroServices.PermissionService
	TrailService      retroServices.TrailService
}

// Routes for Tasks
func (ctrl TaskController) Routes(r *gin.RouterGroup) {
	r.GET("/", ctrl.List)
	r.GET("/:taskID/", ctrl.Get)
	r.GET("/:taskID/members/", ctrl.GetMembers)
	r.POST("/:taskID/members/", ctrl.AddMember)
	r.PUT("/:taskID/members/:smtID/", ctrl.UpdateTaskMember)
}

// List ...
func (ctrl TaskController) List(c *gin.Context) {
	retroID := c.Param("retroID")
	sprintID := c.Param("sprintID")
	userID, _ := c.Get("userID")

	if !ctrl.PermissionService.UserCanAccessSprint(retroID, sprintID, userID.(uint)) {
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
	id := c.Param("taskID")
	retroID := c.Param("retroID")
	sprintID := c.Param("sprintID")
	userID, _ := c.Get("userID")

	if !ctrl.PermissionService.UserCanAccessTask(retroID, sprintID, id, userID.(uint)) {
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
	id := c.Param("taskID")
	retroID := c.Param("retroID")
	sprintID := c.Param("sprintID")
	userID, _ := c.Get("userID")

	if !ctrl.PermissionService.UserCanAccessTask(retroID, sprintID, id, userID.(uint)) {
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

// AddMember adds a member for a task in a particular sprint of a retro
func (ctrl TaskController) AddMember(c *gin.Context) {
	taskID := c.Param("taskID")
	retroID := c.Param("retroID")
	sprintID := c.Param("sprintID")
	userID, _ := c.Get("userID")

	if !ctrl.PermissionService.UserCanEditTask(retroID, sprintID, taskID, userID.(uint)) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{})
		return
	}

	addTaskMemberData := retroSerializers.AddTaskMemberSerializer{}
	if err := c.BindJSON(&addTaskMemberData); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "Invalid request data", "error": err.Error()})
		return
	}

	members, err := ctrl.TaskService.AddMember(taskID, retroID, sprintID, addTaskMemberData.MemberID)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "Couldn't add member", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, members)
}

// UpdateTaskMember updates a member for a task in a particular sprint of a retro
func (ctrl TaskController) UpdateTaskMember(c *gin.Context) {
	taskID := c.Param("taskID")
	retroID := c.Param("retroID")
	sprintID := c.Param("sprintID")
	smtID := c.Param("smtID")
	userID, _ := c.Get("userID")

	if !ctrl.PermissionService.UserCanEditTask(retroID, sprintID, taskID, userID.(uint)) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{})
		return
	}

	taskMemberData := retroSerializers.SprintTaskMemberUpdate{}
	if err := c.BindJSON(&taskMemberData); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "Invalid request data", "error": err.Error()})
		return
	}

	taskMember, err := ctrl.TaskService.UpdateTaskMember(taskID, retroID, sprintID, smtID, &taskMemberData)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "Could not update member", "error": err.Error()})
		return
	}

	ctrl.TrailService.Add("Updated Task Member", "Sprint Task Member", strconv.Itoa(int(taskMember.ID)), userID.(uint))

	c.JSON(http.StatusOK, taskMember)
}
