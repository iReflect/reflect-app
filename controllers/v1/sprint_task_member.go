package v1

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	retroSerializers "github.com/iReflect/reflect-app/apps/retrospective/serializers"
	retroServices "github.com/iReflect/reflect-app/apps/retrospective/services"
	"github.com/iReflect/reflect-app/constants"
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
		responseError := constants.APIErrorMessages[constants.UserCanAccessSprintTaskError]
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": responseError.Message, "code": responseError.Code})
		return
	}

	members, status, errorCode, err := ctrl.SprintTaskMemberService.GetMembers(id, retroID, sprintID)

	if err != nil {
		c.AbortWithStatusJSON(status, gin.H{"error": err.Error(), "code": errorCode})
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
		responseError := constants.APIErrorMessages[constants.UserCanEditSprintTaskError]
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": responseError.Message, "code": responseError.Code})
		return
	}

	addTaskMemberData := retroSerializers.AddSprintTaskMemberSerializer{}
	if err := c.BindJSON(&addTaskMemberData); err != nil {
		responseError := constants.APIErrorMessages[constants.InvalidRequestDataError]
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": responseError.Message, "code": responseError.Code})
		return
	}

	members, status, errorCode, err := ctrl.SprintTaskMemberService.AddMember(sprintTaskID, retroID, sprintID, addTaskMemberData.MemberID)

	if err != nil {
		c.AbortWithStatusJSON(status, gin.H{"error": err.Error(), "code": errorCode})
		return
	}

	ctrl.TrailService.Add(
		constants.AddedSprintMemberTask,
		constants.SprintMemberTask,
		sprintTaskID,
		userID.(uint))

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
		responseError := constants.APIErrorMessages[constants.UserCanEditSprintTaskError]
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": responseError.Message, "code": responseError.Code})
		return
	}

	taskMemberData := retroSerializers.SprintTaskMemberUpdate{}
	if err := c.BindJSON(&taskMemberData); err != nil {
		responseError := constants.APIErrorMessages[constants.InvalidRequestDataError]
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": responseError.Message, "code": responseError.Code})
		return
	}

	taskMember, status, errorCode, err := ctrl.SprintTaskMemberService.UpdateTaskMember(sprintTaskID, retroID, sprintID, smtID, &taskMemberData)

	if err != nil {
		c.AbortWithStatusJSON(status, gin.H{"error": err.Error(), "code": errorCode})
		return
	}

	ctrl.TrailService.Add(
		constants.UpdatedSprintMemberTask,
		constants.SprintMemberTask,
		strconv.Itoa(int(taskMember.ID)),
		userID.(uint))

	c.JSON(status, taskMember)
}
