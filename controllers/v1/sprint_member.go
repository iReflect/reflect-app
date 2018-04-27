package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
	retroSerializers "github.com/iReflect/reflect-app/apps/retrospective/serializers"
	retrospectiveServices "github.com/iReflect/reflect-app/apps/retrospective/services"
	"strconv"
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
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid request data"})
		return
	}

	if !ctrl.PermissionService.UserCanEditSprint(retroID, sprintID, userID.(uint)) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{})
		return
	}

	response, status, err := ctrl.SprintService.AddSprintMember(sprintID, addMemberData.MemberID)
	if err != nil {
		c.AbortWithStatusJSON(status, gin.H{"error": err.Error()})
		return
	}

	ctrl.TrailService.Add("Added Sprint Member", "Sprint Member", strconv.Itoa(int(response.ID)), userID.(uint))

	c.JSON(status, response)
}

// GetSprintMemberList returns the sprint member list
func (ctrl SprintMemberController) GetSprintMemberList(c *gin.Context) {
	userID, _ := c.Get("userID")
	sprintID := c.Param("sprintID")
	retroID := c.Param("retroID")

	if !ctrl.PermissionService.UserCanAccessSprint(retroID, sprintID, userID.(uint)) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{})
		return
	}

	response, status, err := ctrl.SprintService.GetSprintMemberList(sprintID)
	if err != nil {
		c.AbortWithStatusJSON(status, gin.H{"error": err.Error()})
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
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{})
		return
	}

	status, err := ctrl.SprintService.RemoveSprintMember(sprintID, memberID)
	if err != nil {
		c.AbortWithStatusJSON(status, gin.H{"error": err.Error()})
		return
	}

	ctrl.TrailService.Add("Removed", "Sprint Membere", memberID, userID.(uint))

	c.JSON(status, nil)
}

// UpdateSprintMember update the sprint member summary
func (ctrl SprintMemberController) UpdateSprintMember(c *gin.Context) {
	userID, _ := c.Get("userID")
	sprintID := c.Param("sprintID")
	retroID := c.Param("retroID")
	sprintMemberID := c.Param("memberID")

	if !ctrl.PermissionService.UserCanEditSprint(retroID, sprintID, userID.(uint)) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{})
		return
	}
	var memberData retroSerializers.SprintMemberUpdate
	err := c.BindJSON(&memberData)

	response, status, err := ctrl.SprintService.UpdateSprintMember(sprintID, sprintMemberID, memberData)
	if err != nil {
		c.AbortWithStatusJSON(status, gin.H{"error": err.Error()})
		return
	}

	ctrl.TrailService.Add("Updated Sprint Member", "Sprint Member", sprintMemberID, userID.(uint))

	c.JSON(status, response)
}
