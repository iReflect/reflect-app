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

// SprintHighlightController ...
type SprintHighlightController struct {
	RetrospectiveFeedbackService retrospectiveServices.RetrospectiveFeedbackService
	PermissionService            retrospectiveServices.PermissionService
	TrailService                 retrospectiveServices.TrailService
}

// Routes for Sprints
func (ctrl SprintHighlightController) Routes(r *gin.RouterGroup) {
	r.POST("/", ctrl.Add)
	r.GET("/", ctrl.List)
	r.DELETE("/:highlightID/", ctrl.Delete)
	r.PATCH("/:highlightID/", ctrl.Update)
}

// Add Highlight to sprint's retrospective
func (ctrl SprintHighlightController) Add(c *gin.Context) {
	userID, _ := c.Get("userID")
	sprintID := c.Param("sprintID")
	retroID := c.Param("retroID")
	feedbackData := serializers.RetrospectiveFeedbackCreateSerializer{}

	if err := c.BindJSON(&feedbackData); err != nil {
		responseError := constants.APIErrorMessages[constants.InvalidRequestDataError]
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": responseError.Message, "code": responseError.Code})
		return
	}

	if !ctrl.PermissionService.CanAccessRetrospectiveFeedback(sprintID, userID.(uint)) {
		responseError := constants.APIErrorMessages[constants.RetrospectiveFeedbackAccessError]
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": responseError.Message, "code": responseError.Code})
		return
	}

	if !ctrl.PermissionService.UserCanEditSprint(retroID, sprintID, userID.(uint)) {
		responseError := constants.APIErrorMessages[constants.UserCanEditSprintError]
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": responseError.Message, "code": responseError.Code})
		return
	}

	response, status, errorCode, err := ctrl.RetrospectiveFeedbackService.Add(
		userID.(uint),
		sprintID,
		retroID,
		models.HighlightType,
		&feedbackData)
	if err != nil {
		c.AbortWithStatusJSON(status, gin.H{"error": err.Error(), "code": errorCode})
		return
	}

	ctrl.TrailService.Add(
		constants.AddedHighlight,
		constants.RetrospectiveFeedback,
		fmt.Sprint(response.ID),
		userID.(uint))

	c.JSON(status, response)
}

// List highlights associated to sprint
func (ctrl SprintHighlightController) List(c *gin.Context) {
	userID, _ := c.Get("userID")
	sprintID := c.Param("sprintID")
	retroID := c.Param("retroID")

	if !ctrl.PermissionService.CanAccessRetrospectiveFeedback(sprintID, userID.(uint)) {
		responseError := constants.APIErrorMessages[constants.RetrospectiveFeedbackAccessError]
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": responseError.Message, "code": responseError.Code})
		return
	}

	if !ctrl.PermissionService.UserCanAccessSprint(retroID, sprintID, userID.(uint)) {
		responseError := constants.APIErrorMessages[constants.UserCanAccessSprintError]
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": responseError.Message, "code": responseError.Code})
		return
	}

	response, status, errorCode, err := ctrl.RetrospectiveFeedbackService.List(
		userID.(uint),
		sprintID,
		retroID,
		models.HighlightType)
	if err != nil {
		c.AbortWithStatusJSON(status, gin.H{"error": err.Error(), "code": errorCode})
		return
	}

	c.JSON(status, response)
}

// Update highlight associated to a sprint
func (ctrl SprintHighlightController) Update(c *gin.Context) {
	userID, _ := c.Get("userID")
	sprintID := c.Param("sprintID")
	retroID := c.Param("retroID")
	highlightID := c.Param("highlightID")
	feedbackData := serializers.RetrospectiveFeedbackUpdateSerializer{}

	if err := c.BindJSON(&feedbackData); err != nil {
		responseError := constants.APIErrorMessages[constants.InvalidRequestDataError]
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": responseError.Message, "code": responseError.Code})
		return
	}

	if !ctrl.PermissionService.CanAccessRetrospectiveFeedback(sprintID, userID.(uint)) {
		responseError := constants.APIErrorMessages[constants.RetrospectiveFeedbackAccessError]
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": responseError.Message, "code": responseError.Code})
		return
	}

	if !ctrl.PermissionService.UserCanEditSprint(retroID, sprintID, userID.(uint)) {
		responseError := constants.APIErrorMessages[constants.UserCanEditSprintError]
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": responseError.Message, "code": responseError.Code})
		return
	}

	response, status, errorCode, err := ctrl.RetrospectiveFeedbackService.Update(
		userID.(uint),
		retroID,
		highlightID,
		&feedbackData)
	if err != nil {
		c.AbortWithStatusJSON(status, gin.H{"error": err.Error(), "code": errorCode})
		return
	}

	ctrl.TrailService.Add(
		constants.UpdatedHighlight,
		constants.RetrospectiveFeedback,
		highlightID,
		userID.(uint))

	c.JSON(status, response)
}

// Delete ...
func (ctrl SprintHighlightController) Delete(c *gin.Context) {
	userID, _ := c.Get("userID")
	sprintID := c.Param("sprintID")
	retroID := c.Param("retroID")
	highlightID := c.Param("highlightID")

	if !ctrl.PermissionService.CanAccessRetrospectiveFeedback(sprintID, userID.(uint)) {
		responseError := constants.APIErrorMessages[constants.RetrospectiveFeedbackAccessError]
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": responseError.Message, "code": responseError.Code})
		return
	}
	if !ctrl.PermissionService.UserCanEditSprint(retroID, sprintID, userID.(uint)) {
		responseError := constants.APIErrorMessages[constants.UserCanEditSprintError]
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": responseError.Message, "code": responseError.Code})
		return
	}
	status, err := ctrl.RetrospectiveFeedbackService.Delete(highlightID)
	if err != nil {
		responseError := constants.APIErrorMessages[constants.DeleteRetroFeedbackHighlightError]
		c.AbortWithStatusJSON(status, gin.H{"error": responseError.Message, "code": responseError.Code})
	}
	ctrl.TrailService.Add(
		constants.DeletedHighlight,
		constants.RetrospectiveFeedback,
		highlightID,
		userID.(uint))

	c.JSON(status, nil)
}
