package v1

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	retrospectiveSerializers "github.com/iReflect/reflect-app/apps/retrospective/serializers"
	retrospectiveService "github.com/iReflect/reflect-app/apps/retrospective/services"
)

// RetroSpectiveController ...
type RetrospectiveController struct {
	RetrospectiveService retrospectiveService.RetrospectiveService
	PermissionService    retrospectiveService.PermissionService
}

// Routes for RetroSpective
func (ctrl RetrospectiveController) Routes(r *gin.RouterGroup) {
	r.GET("/", ctrl.List)
	r.GET("/:retroID", ctrl.Get)
	r.GET(":retroID/latest-sprint", ctrl.GetLatestSprint)
	r.POST("/", ctrl.Create)
}

// List RetroSpectives
func (ctrl RetrospectiveController) List(c *gin.Context) {
	userID, _ := c.Get("userID")
	perPage, err := strconv.Atoi(c.DefaultQuery("perPage", ""))
	if err != nil {
		perPage = -1
	}
	page, err := strconv.Atoi(c.DefaultQuery("page", ""))
	if err != nil {
		page = 1
	}
	response, err := ctrl.RetrospectiveService.List(userID.(uint), perPage, page)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "Retrospectives not found", "error": err})
	}
	c.JSON(http.StatusOK, response)
}

// Get RetroSpective by id
func (ctrl RetrospectiveController) Get(c *gin.Context) {
	retroID := c.Param("retroID")
	userID, _ := c.Get("userID")

	if !ctrl.PermissionService.UserCanAccessRetro(retroID, userID.(uint)) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{})
		return
	}
	response, err := ctrl.RetrospectiveService.Get(retroID)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "Retrospective not found", "error": err})
		return
	}
	c.JSON(http.StatusOK, response)
}

// GetLatestSprint returns the latest active/freezed sprint's data
func (ctrl RetrospectiveController) GetLatestSprint(c *gin.Context) {
	userID, _ := c.Get("userID")
	retroID := c.Param("retroID")
	if !ctrl.PermissionService.UserCanAccessRetro(retroID, userID.(uint)) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{})
		return
	}

	sprint, err := ctrl.RetrospectiveService.GetLatestSprint(retroID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "Failed to get latest sprint data", "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, sprint)

}

// Create RetroSpective
func (ctrl RetrospectiveController) Create(c *gin.Context) {
	userID, _ := c.Get("userID")
	var err error
	retrospectiveData := retrospectiveSerializers.RetrospectiveCreateSerializer{CreatedByID: userID.(uint)}
	if err = c.BindJSON(&retrospectiveData); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "Invalid request data", "error": err.Error()})
		return
	}
	err = ctrl.RetrospectiveService.Create(userID.(uint), &retrospectiveData)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest,
			gin.H{"message": "Retrospective can't be created", "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}
