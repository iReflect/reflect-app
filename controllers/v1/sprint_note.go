package v1

import (
	"fmt"
	"net/http"

	"github.com/iReflect/reflect-app/apps/retrospective/models"
	"github.com/iReflect/reflect-app/apps/retrospective/serializers"
	"github.com/iReflect/reflect-app/constants"

	"github.com/gin-gonic/gin"
	retrospectiveServices "github.com/iReflect/reflect-app/apps/retrospective/services"
)

// SprintNoteController ...
type SprintNoteController struct {
	RetrospectiveFeedbackService retrospectiveServices.RetrospectiveFeedbackService
	PermissionService            retrospectiveServices.PermissionService
	TrailService                 retrospectiveServices.TrailService
}

// Routes for Sprints
func (ctrl SprintNoteController) Routes(r *gin.RouterGroup) {
	r.POST("/", ctrl.Add)
	r.GET("/", ctrl.List)
	r.PATCH("/:noteID/", ctrl.Update)
}

// Add Note to sprint's retrospective
func (ctrl SprintNoteController) Add(c *gin.Context) {
	userID, _ := c.Get("userID")
	sprintID := c.Param("sprintID")
	retroID := c.Param("retroID")
	feedbackData := serializers.RetrospectiveFeedbackCreateSerializer{}

	if err := c.BindJSON(&feedbackData); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid request data"})
		return
	}

	if !ctrl.PermissionService.CanAccessRetrospectiveFeedback(sprintID) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{})
		return
	}

	if !ctrl.PermissionService.UserCanEditSprint(retroID, sprintID, userID.(uint)) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{})
		return
	}

	response, status, err := ctrl.RetrospectiveFeedbackService.Add(
		userID.(uint),
		sprintID,
		retroID,
		models.NoteType,
		&feedbackData)
	if err != nil {
		c.AbortWithStatusJSON(status, gin.H{"error": err.Error()})
		return
	}

	ctrl.TrailService.Add(
		constants.AddedNote,
		constants.RetrospectiveFeedback,
		fmt.Sprint(response.ID),
		userID.(uint))

	c.JSON(status, response)
}

// List notes associated to sprint
func (ctrl SprintNoteController) List(c *gin.Context) {
	userID, _ := c.Get("userID")
	sprintID := c.Param("sprintID")
	retroID := c.Param("retroID")

	if !ctrl.PermissionService.CanAccessRetrospectiveFeedback(sprintID) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{})
		return
	}

	if !ctrl.PermissionService.UserCanAccessSprint(retroID, sprintID, userID.(uint)) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{})
		return
	}

	response, status, err := ctrl.RetrospectiveFeedbackService.List(
		userID.(uint),
		sprintID,
		retroID,
		models.NoteType)
	if err != nil {
		c.AbortWithStatusJSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(status, response)
}

// Update note associated to a sprint
func (ctrl SprintNoteController) Update(c *gin.Context) {
	userID, _ := c.Get("userID")
	sprintID := c.Param("sprintID")
	retroID := c.Param("retroID")
	noteID := c.Param("noteID")
	feedbackData := serializers.RetrospectiveFeedbackUpdateSerializer{}

	if err := c.BindJSON(&feedbackData); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid request data"})
		return
	}

	if !ctrl.PermissionService.CanAccessRetrospectiveFeedback(sprintID) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{})
		return
	}

	if !ctrl.PermissionService.UserCanEditSprint(retroID, sprintID, userID.(uint)) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{})
		return
	}

	response, status, err := ctrl.RetrospectiveFeedbackService.Update(
		userID.(uint),
		retroID,
		noteID,
		&feedbackData)
	if err != nil {
		c.AbortWithStatusJSON(status, gin.H{"error": err.Error()})
		return
	}

	ctrl.TrailService.Add(
		constants.UpdatedNote,
		constants.RetrospectiveFeedback,
		noteID,
		userID.(uint))

	c.JSON(status, response)
}
