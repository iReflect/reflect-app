package v1

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/iReflect/reflect-app/apps/retrospective/models"
	"github.com/iReflect/reflect-app/apps/retrospective/serializers"
	retrospectiveServices "github.com/iReflect/reflect-app/apps/retrospective/services"
	"github.com/iReflect/reflect-app/constants"
)

// SprintGoalController ...
type SprintGoalController struct {
	RetrospectiveFeedbackService retrospectiveServices.RetrospectiveFeedbackService
	PermissionService            retrospectiveServices.PermissionService
	TrailService                 retrospectiveServices.TrailService
}

// Routes for Sprints
func (ctrl SprintGoalController) Routes(r *gin.RouterGroup) {
	r.POST("/", ctrl.Add)
	r.GET("/", ctrl.List)
	r.PATCH("/:goalID/", ctrl.Update)
	r.POST("/:goalID/resolve/", ctrl.Resolve)
	r.DELETE("/:goalID/resolve/", ctrl.UnResolve)
}

// Add Goal to sprint's retrospective
func (ctrl SprintGoalController) Add(c *gin.Context) {
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
		models.GoalType,
		&feedbackData,
	)
	if err != nil {
		c.AbortWithStatusJSON(status, gin.H{"error": err.Error()})
		return
	}

	ctrl.TrailService.Add(
		constants.AddedGoal,
		constants.RetrospectiveFeedback,
		fmt.Sprint(response.ID),
		userID.(uint))
	c.JSON(status, response)
}

// List goals associated to sprint
func (ctrl SprintGoalController) List(c *gin.Context) {
	userID, _ := c.Get("userID")
	sprintID := c.Param("sprintID")
	retroID := c.Param("retroID")
	goalType := c.Query("goalType")

	if !ctrl.PermissionService.CanAccessRetrospectiveFeedback(sprintID) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{})
		return
	}

	if !ctrl.PermissionService.UserCanAccessSprint(retroID, sprintID, userID.(uint)) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{})
		return
	}

	response, status, err := ctrl.RetrospectiveFeedbackService.ListGoal(
		userID.(uint),
		sprintID,
		retroID,
		goalType)
	if err != nil {
		c.AbortWithStatusJSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(status, response)
}

// Update goal associated to a sprint
func (ctrl SprintGoalController) Update(c *gin.Context) {
	userID, _ := c.Get("userID")
	sprintID := c.Param("sprintID")
	retroID := c.Param("retroID")
	goalID := c.Param("goalID")

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
		goalID,
		&feedbackData)
	if err != nil {
		c.AbortWithStatusJSON(status, gin.H{"error": err.Error()})
		return
	}

	ctrl.TrailService.Add(
		constants.UpdatedGoal,
		constants.RetrospectiveFeedback,
		goalID,
		userID.(uint))

	c.JSON(status, response)
}

// Resolve goal associated to a sprint
func (ctrl SprintGoalController) Resolve(c *gin.Context) {
	userID, _ := c.Get("userID")
	sprintID := c.Param("sprintID")
	retroID := c.Param("retroID")
	goalID := c.Param("goalID")

	if !ctrl.PermissionService.CanAccessRetrospectiveFeedback(sprintID) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{})
		return
	}

	if !ctrl.PermissionService.UserCanEditSprint(retroID, sprintID, userID.(uint)) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{})
		return
	}

	response, status, err := ctrl.RetrospectiveFeedbackService.Resolve(
		userID.(uint),
		sprintID,
		retroID,
		goalID,
		true)
	if err != nil {
		c.AbortWithStatusJSON(status, gin.H{"error": err.Error()})
		return
	}
	ctrl.TrailService.Add(
		constants.ResolvedGoal,
		constants.RetrospectiveFeedback,
		goalID,
		userID.(uint))

	c.JSON(status, response)
}

// UnResolve a goal associated to a sprint
func (ctrl SprintGoalController) UnResolve(c *gin.Context) {
	userID, _ := c.Get("userID")
	sprintID := c.Param("sprintID")
	retroID := c.Param("retroID")
	goalID := c.Param("goalID")

	if !ctrl.PermissionService.CanAccessRetrospectiveFeedback(sprintID) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{})
		return
	}

	if !ctrl.PermissionService.UserCanEditSprint(retroID, sprintID, userID.(uint)) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{})
		return
	}

	response, status, err := ctrl.RetrospectiveFeedbackService.Resolve(
		userID.(uint),
		sprintID,
		retroID,
		goalID,
		false)
	if err != nil {
		c.AbortWithStatusJSON(status, gin.H{"error": err.Error()})
		return
	}

	ctrl.TrailService.Add(
		constants.UnresolvedGoal,
		constants.RetrospectiveFeedback,
		goalID,
		userID.(uint))

	c.JSON(status, response)
}
