package v1

import (
	"github.com/iReflect/reflect-app/apps/retrospective/models"
	"github.com/iReflect/reflect-app/apps/retrospective/serializers"
	"net/http"

	"github.com/gin-gonic/gin"
	retrospectiveServices "github.com/iReflect/reflect-app/apps/retrospective/services"
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
	r.PUT("/:goalID/", ctrl.Update)
	r.POST("/:goalID/resolve", ctrl.Resolve)
	r.DELETE("/:goalID/resolve", ctrl.UnResolve)
}

// Add Goal to sprint's retrospective
func (ctrl SprintGoalController) Add(c *gin.Context) {
	userID, _ := c.Get("userID")
	sprintID := c.Param("sprintID")
	retroID := c.Param("retroID")
	feedbackData := serializers.RetrospectiveFeedbackCreateSerializer{}

	if err := c.BindJSON(&feedbackData); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "Invalid request data", "error": err.Error()})
		return
	}

	if !ctrl.PermissionService.UserCanEditSprint(retroID, sprintID, userID.(uint)) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{})
		return
	}

	response, err := ctrl.RetrospectiveFeedbackService.Add(
		userID.(uint),
		sprintID,
		retroID,
		models.GoalType,
		&feedbackData)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "Failed to create goal",
			"error":   err.Error()})
		return
	}

	c.JSON(http.StatusCreated, response)
}

// List goals associated to sprint
func (ctrl SprintGoalController) List(c *gin.Context) {
	userID, _ := c.Get("userID")
	sprintID := c.Param("sprintID")
	retroID := c.Param("retroID")
	goalType := c.Query("goalType")

	if !ctrl.PermissionService.UserCanAccessSprint(retroID, sprintID, userID.(uint)) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{})
		return
	}
	response, err := ctrl.RetrospectiveFeedbackService.ListGoal(
		userID.(uint),
		sprintID,
		retroID,
		goalType)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "Failed to fetch goals",
			"error":   err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// Update goal associated to a sprint
func (ctrl SprintGoalController) Update(c *gin.Context) {
	userID, _ := c.Get("userID")
	sprintID := c.Param("sprintID")
	retroID := c.Param("retroID")
	goalID := c.Param("goalID")
	feedbackData := serializers.RetrospectiveFeedbackUpdateSerializer{}

	if err := c.BindJSON(&feedbackData); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "Invalid request data", "error": err.Error()})
		return
	}

	if !ctrl.PermissionService.UserCanEditSprint(retroID, sprintID, userID.(uint)) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{})
		return
	}

	response, err := ctrl.RetrospectiveFeedbackService.Update(
		userID.(uint),
		sprintID,
		retroID,
		goalID,
		&feedbackData)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "Failed to update goal",
			"error":   err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// Resolve goal associated to a sprint
func (ctrl SprintGoalController) Resolve(c *gin.Context) {
	userID, _ := c.Get("userID")
	sprintID := c.Param("sprintID")
	retroID := c.Param("retroID")
	goalID := c.Param("goalID")

	if !ctrl.PermissionService.UserCanEditSprint(retroID, sprintID, userID.(uint)) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{})
		return
	}

	response, err := ctrl.RetrospectiveFeedbackService.Resolve(
		userID.(uint),
		sprintID,
		retroID,
		goalID,
		true)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "Failed to resolved goal",
			"error":   err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// UnResolve a goal associated to a sprint
func (ctrl SprintGoalController) UnResolve(c *gin.Context) {
	userID, _ := c.Get("userID")
	sprintID := c.Param("sprintID")
	retroID := c.Param("retroID")
	goalID := c.Param("goalID")

	if !ctrl.PermissionService.UserCanEditSprint(retroID, sprintID, userID.(uint)) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{})
		return
	}

	response, err := ctrl.RetrospectiveFeedbackService.Resolve(
		userID.(uint),
		sprintID,
		retroID,
		goalID,
		false)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "Failed to un-resolve goal",
			"error":   err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}
