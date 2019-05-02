package v1

import (
	"net/http"

	"strconv"

	"github.com/gin-gonic/gin"
	retroSerializers "github.com/iReflect/reflect-app/apps/retrospective/serializers"
	retrospectiveServices "github.com/iReflect/reflect-app/apps/retrospective/services"
	"github.com/iReflect/reflect-app/constants"
)

// SprintMemberController ...
type SprintMemberController struct {
	SprintService     retrospectiveServices.SprintService
	PermissionService retrospectiveServices.PermissionService
	TrailService      retrospectiveServices.TrailService
}

// Routes for Sprints
func (ctrl SprintMemberController) Routes(r *gin.RouterGroup) {
	r.POST("/", ctrl.AddMember)
	r.GET("/", ctrl.GetSprintMemberList)
	r.PATCH("/:memberID/", ctrl.UpdateSprintMember)
	r.DELETE("/:memberID/", ctrl.RemoveMember)
}

// AddMember to a Sprint
func (ctrl SprintMemberController) AddMember(c *gin.Context) {
	userID, _ := c.Get("userID")
	sprintID := c.Param("sprintID")
	retroID := c.Param("retroID")

	addMemberData := retroSerializers.AddMemberSerializer{}
	if err := c.BindJSON(&addMemberData); err != nil {
		responseError := constants.APIErrorMessages[constants.InvalidRequestDataError]
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": responseError.Message, "code": responseError.Code})
		return
	}

	if !ctrl.PermissionService.UserCanEditSprint(retroID, sprintID, userID.(uint)) {
		responseError := constants.APIErrorMessages[constants.UserCanEditSprintError]
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": responseError.Message, "code": responseError.Code})
		return
	}

	response, status, errorCode, err := ctrl.SprintService.AddSprintMember(sprintID, addMemberData.MemberID)
	if err != nil {
		c.AbortWithStatusJSON(status, gin.H{"error": err.Error(), "code": errorCode})
		return
	}

	ctrl.TrailService.Add(
		constants.AddedSprintMember,
		constants.SprintMember,
		strconv.Itoa(int(response.ID)),
		userID.(uint))

	c.JSON(status, response)
}

// GetSprintMemberList returns the sprint member list
func (ctrl SprintMemberController) GetSprintMemberList(c *gin.Context) {
	userID, _ := c.Get("userID")
	sprintID := c.Param("sprintID")
	retroID := c.Param("retroID")

	if !ctrl.PermissionService.UserCanAccessSprint(retroID, sprintID, userID.(uint)) {
		responseError := constants.APIErrorMessages[constants.UserCanAccessSprintError]
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": responseError.Message, "code": responseError.Code})
		return
	}

	response, status, errorCode, err := ctrl.SprintService.GetSprintMemberList(sprintID)
	if err != nil {
		c.AbortWithStatusJSON(status, gin.H{"error": err.Error(), "code": errorCode})
		return
	}

	c.JSON(status, response)
}

// RemoveMember from a Sprint
func (ctrl SprintMemberController) RemoveMember(c *gin.Context) {
	userID, _ := c.Get("userID")
	sprintID := c.Param("sprintID")
	retroID := c.Param("retroID")
	memberID := c.Param("memberID")

	if !ctrl.PermissionService.UserCanEditSprint(retroID, sprintID, userID.(uint)) {
		responseError := constants.APIErrorMessages[constants.UserCanEditSprintError]
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": responseError.Message, "code": responseError.Code})
		return
	}

	status, errorCode, err := ctrl.SprintService.RemoveSprintMember(sprintID, memberID)
	if err != nil {
		c.AbortWithStatusJSON(status, gin.H{"error": err.Error(), "code": errorCode})
		return
	}

	ctrl.TrailService.Add(
		constants.RemovedSprintMember,
		constants.SprintMember,
		memberID,
		userID.(uint))

	c.JSON(status, nil)
}

// UpdateSprintMember update the sprint member summary
func (ctrl SprintMemberController) UpdateSprintMember(c *gin.Context) {
	userID, _ := c.Get("userID")
	sprintID := c.Param("sprintID")
	retroID := c.Param("retroID")
	sprintMemberID := c.Param("memberID")

	if !ctrl.PermissionService.UserCanEditSprint(retroID, sprintID, userID.(uint)) {
		responseError := constants.APIErrorMessages[constants.UserCanEditSprintError]
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": responseError.Message, "code": responseError.Code})
		return
	}
	var memberData retroSerializers.SprintMemberUpdate
	if err := c.BindJSON(&memberData); err != nil {
		responseError := constants.APIErrorMessages[constants.InvalidRequestDataError]
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": responseError.Message, "code": responseError.Code})
		return
	}

	response, status, errorCode, err := ctrl.SprintService.UpdateSprintMember(sprintID, sprintMemberID, memberData)
	if err != nil {
		c.AbortWithStatusJSON(status, gin.H{"error": err.Error(), "code": errorCode})
		return
	}

	ctrl.TrailService.Add(
		constants.UpdatedSprintMember,
		constants.SprintMember,
		sprintMemberID,
		userID.(uint))

	c.JSON(status, response)
}
