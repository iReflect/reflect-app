package v1

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	retrospectiveSerializers "github.com/iReflect/reflect-app/apps/retrospective/serializers"
	retrospectiveService "github.com/iReflect/reflect-app/apps/retrospective/services"
)

// RetrospectiveController ...
type RetrospectiveController struct {
	RetrospectiveService retrospectiveService.RetrospectiveService
	PermissionService    retrospectiveService.PermissionService
	TrailService         retrospectiveService.TrailService
}

// Routes for Retrospective
func (ctrl RetrospectiveController) Routes(r *gin.RouterGroup) {
	r.GET("/", ctrl.List)
	r.GET("/:retroID/", ctrl.Get)
	r.GET(":retroID/latest-sprint/", ctrl.GetLatestSprint)
	r.POST("/", ctrl.Create)
}

// List Retrospectives
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
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": err.Error()})
	}
	c.JSON(http.StatusOK, response)
}

// Get Retrospective by id
func (ctrl RetrospectiveController) Get(c *gin.Context) {
	retroID := c.Param("retroID")
	userID, _ := c.Get("userID")

	if !ctrl.PermissionService.UserCanAccessRetro(retroID, userID.(uint)) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{})
		return
	}
	response, err := ctrl.RetrospectiveService.Get(retroID)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, response)
}

// GetLatestSprint returns the latest active/frozen sprint's data
func (ctrl RetrospectiveController) GetLatestSprint(c *gin.Context) {
	userID, _ := c.Get("userID")
	retroID := c.Param("retroID")
	if !ctrl.PermissionService.UserCanAccessRetro(retroID, userID.(uint)) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{})
		return
	}

	sprint, err := ctrl.RetrospectiveService.GetLatestSprint(retroID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, sprint)

}

// Create Retrospective
func (ctrl RetrospectiveController) Create(c *gin.Context) {
	userID, _ := c.Get("userID")
	var err error
	retrospectiveData := retrospectiveSerializers.RetrospectiveCreateSerializer{CreatedByID: userID.(uint)}
	if err = c.BindJSON(&retrospectiveData); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}
	retro, err := ctrl.RetrospectiveService.Create(userID.(uint), &retrospectiveData)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctrl.TrailService.Add("Created Retrospective", "Retrospective", strconv.Itoa(int(retro.ID)), userID.(uint))

	c.JSON(http.StatusCreated, retro)
}
