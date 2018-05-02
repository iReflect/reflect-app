package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
	retroSerializers "github.com/iReflect/reflect-app/apps/retrospective/serializers"
	retroServices "github.com/iReflect/reflect-app/apps/retrospective/services"
	"strconv"
)

// SprinTaskController ...
type SprinTaskController struct {
	SprintTaskService retroServices.SprintTaskService
	PermissionService retroServices.PermissionService
	TrailService      retroServices.TrailService
}

// Routes for Tasks
func (ctrl SprinTaskController) Routes(r *gin.RouterGroup) {
	r.GET("/", ctrl.List)
	r.GET("/:sprintTaskID/", ctrl.Get)
	r.PATCH("/:sprintTaskID/", ctrl.Update)
	r.POST("/:sprintTaskID/done/", ctrl.MarkDone)
	r.DELETE("/:sprintTaskID/done/", ctrl.MarkUndone)
	r.GET("/:sprintTaskID/members/", ctrl.GetMembers)
	r.POST("/:sprintTaskID/members/", ctrl.AddMember)
	r.PATCH("/:sprintTaskID/members/:smtID/", ctrl.UpdateTaskMember)
}

// List ...
func (ctrl SprinTaskController) List(c *gin.Context) {
	retroID := c.Param("retroID")
	sprintID := c.Param("sprintID")
	userID, _ := c.Get("userID")

	if !ctrl.PermissionService.UserCanAccessSprint(retroID, sprintID, userID.(uint)) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{})
		return
	}

	tasks, status, err := ctrl.SprintTaskService.List(retroID, sprintID)

	if err != nil {
		c.AbortWithStatusJSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(status, tasks)
}

// Get ...
func (ctrl SprinTaskController) Get(c *gin.Context) {
	id := c.Param("sprintTaskID")
	retroID := c.Param("retroID")
	sprintID := c.Param("sprintID")
	userID, _ := c.Get("userID")

	if !ctrl.PermissionService.UserCanAccessSprintTask(retroID, sprintID, id, userID.(uint)) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{})
		return
	}

	task, status, err := ctrl.SprintTaskService.Get(id, retroID, sprintID)

	if err != nil {
		c.AbortWithStatusJSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(status, task)
}

// Update ...
func (ctrl SprinTaskController) Update(c *gin.Context) {
	id := c.Param("sprintTaskID")
	retroID := c.Param("retroID")
	sprintID := c.Param("sprintID")
	userID, _ := c.Get("userID")

	if !ctrl.PermissionService.UserCanEditSprintTask(retroID, sprintID, id, userID.(uint)) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{})
		return
	}

	var data retroSerializers.SprintTaskUpdate
	err := c.BindJSON(&data)

	task, status, err := ctrl.SprintTaskService.Update(id, retroID, sprintID, data)

	if err != nil {
		c.AbortWithStatusJSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(status, task)
}

// MarkDone ...
func (ctrl SprinTaskController) MarkDone(c *gin.Context) {
	id := c.Param("sprintTaskID")
	retroID := c.Param("retroID")
	sprintID := c.Param("sprintID")
	userID, _ := c.Get("userID")

	if !ctrl.PermissionService.UserCanEditSprintTask(retroID, sprintID, id, userID.(uint)) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{})
		return
	}

	task, status, err := ctrl.SprintTaskService.MarkDone(id, retroID, sprintID)

	if err != nil {
		c.AbortWithStatusJSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(status, task)
}

// MarkUndone ...
func (ctrl SprinTaskController) MarkUndone(c *gin.Context) {
	id := c.Param("sprintTaskID")
	retroID := c.Param("retroID")
	sprintID := c.Param("sprintID")
	userID, _ := c.Get("userID")

	if !ctrl.PermissionService.UserCanEditSprintTask(retroID, sprintID, id, userID.(uint)) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{})
		return
	}

	task, status, err := ctrl.SprintTaskService.MarkUndone(id, retroID, sprintID)

	if err != nil {
		c.AbortWithStatusJSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(status, task)
}

// GetMembers ...
func (ctrl SprinTaskController) GetMembers(c *gin.Context) {
	id := c.Param("sprintTaskID")
	retroID := c.Param("retroID")
	sprintID := c.Param("sprintID")
	userID, _ := c.Get("userID")

	if !ctrl.PermissionService.UserCanAccessSprintTask(retroID, sprintID, id, userID.(uint)) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{})
		return
	}

	members, status, err := ctrl.SprintTaskService.GetMembers(id, retroID, sprintID)

	if err != nil {
		c.AbortWithStatusJSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(status, members)
}

// AddMember adds a member for a task in a particular sprint of a retro
func (ctrl SprinTaskController) AddMember(c *gin.Context) {
	sprintTaskID := c.Param("sprintTaskID")
	retroID := c.Param("retroID")
	sprintID := c.Param("sprintID")
	userID, _ := c.Get("userID")

	if !ctrl.PermissionService.UserCanEditSprintTask(retroID, sprintID, sprintTaskID, userID.(uint)) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{})
		return
	}

	addTaskMemberData := retroSerializers.AddSprintTaskMemberSerializer{}
	if err := c.BindJSON(&addTaskMemberData); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid request data"})
		return
	}

	members, status, err := ctrl.SprintTaskService.AddMember(sprintTaskID, retroID, sprintID, addTaskMemberData.MemberID)

	if err != nil {
		c.AbortWithStatusJSON(status, gin.H{"error": err.Error()})
		return
	}

	ctrl.TrailService.Add("Added SprintTask Member", "Sprint Member SprintTask", sprintTaskID, userID.(uint))

	c.JSON(status, members)
}

// UpdateTaskMember updates a member for a task in a particular sprint of a retro
func (ctrl SprinTaskController) UpdateTaskMember(c *gin.Context) {
	sprintTaskID := c.Param("sprintTaskID")
	retroID := c.Param("retroID")
	sprintID := c.Param("sprintID")
	smtID := c.Param("smtID")
	userID, _ := c.Get("userID")

	if !ctrl.PermissionService.UserCanEditSprintTask(retroID, sprintID, sprintTaskID, userID.(uint)) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{})
		return
	}

	taskMemberData := retroSerializers.SprintTaskMemberUpdate{}
	if err := c.BindJSON(&taskMemberData); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid request data"})
		return
	}

	taskMember, status, err := ctrl.SprintTaskService.UpdateTaskMember(sprintTaskID, retroID, sprintID, smtID, &taskMemberData)

	if err != nil {
		c.AbortWithStatusJSON(status, gin.H{"error": err.Error()})
		return
	}

	ctrl.TrailService.Add("Updated SprintTask Member", "Sprint Member SprintTask", strconv.Itoa(int(taskMember.ID)), userID.(uint))

	c.JSON(status, taskMember)
}
