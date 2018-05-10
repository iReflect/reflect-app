package v1

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	retroSerializers "github.com/iReflect/reflect-app/apps/retrospective/serializers"
	retroServices "github.com/iReflect/reflect-app/apps/retrospective/services"
)

// SprintTaskMemberController ...
type SprintTaskMemberController struct {
	SprintTaskMemberService retroServices.SprintTaskMemberService
	PermissionService       retroServices.PermissionService
	TrailService            retroServices.TrailService
}

// Routes for Tasks
func (ctrl SprintTaskMemberController) Routes(r *gin.RouterGroup) {
	r.GET("/", ctrl.GetMembers)
	r.POST("/", ctrl.AddMember)
	r.PATCH("/:smtID/", ctrl.UpdateTaskMember)
}

// GetMembers ...
func (ctrl SprintTaskMemberController) GetMembers(c *gin.Context) {
	id := c.Param("sprintTaskID")
	retroID := c.Param("retroID")
	sprintID := c.Param("sprintID")
	userID, _ := c.Get("userID")

	if !ctrl.PermissionService.UserCanAccessSprintTask(retroID, sprintID, id, userID.(uint)) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{})
		return
	}

	members, status, err := ctrl.SprintTaskMemberService.GetMembers(id, retroID, sprintID)

	if err != nil {
		c.AbortWithStatusJSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(status, members)
}

// AddMember adds a member for a task in a particular sprint of a retro
func (ctrl SprintTaskMemberController) AddMember(c *gin.Context) {
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

	members, status, err := ctrl.SprintTaskMemberService.AddMember(sprintTaskID, retroID, sprintID, addTaskMemberData.MemberID)

	if err != nil {
		c.AbortWithStatusJSON(status, gin.H{"error": err.Error()})
		return
	}

	ctrl.TrailService.Add("Added SprintTask Member", "Sprint Member SprintTask", sprintTaskID, userID.(uint))

	c.JSON(status, members)
}

// UpdateTaskMember updates a member for a task in a particular sprint of a retro
func (ctrl SprintTaskMemberController) UpdateTaskMember(c *gin.Context) {
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

	taskMember, status, err := ctrl.SprintTaskMemberService.UpdateTaskMember(sprintTaskID, retroID, sprintID, smtID, &taskMemberData)

	if err != nil {
		c.AbortWithStatusJSON(status, gin.H{"error": err.Error()})
		return
	}

	ctrl.TrailService.Add("Updated SprintTask Member", "Sprint Member SprintTask", strconv.Itoa(int(taskMember.ID)), userID.(uint))

	c.JSON(status, taskMember)
}
