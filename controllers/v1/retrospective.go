package v1

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	retrospectiveSerializers "github.com/iReflect/reflect-app/apps/retrospective/serializers"
	retrospectiveService "github.com/iReflect/reflect-app/apps/retrospective/services"
	"github.com/iReflect/reflect-app/constants"
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
	r.PUT("/:retroID/", ctrl.Update)
	r.GET("/:retroID/edit-level/", ctrl.GetEditLevels)
	r.GET("/:retroID/team-members/", ctrl.GetTeamMembers)
	r.GET("/:retroID/latest-sprint/", ctrl.GetLatestSprint)
	r.POST("/", ctrl.Create)
}

// List Retrospectives
func (ctrl RetrospectiveController) List(c *gin.Context) {
	userID, _ := c.Get("userID")
	perPage := c.DefaultQuery("perPage", "")
	page := c.DefaultQuery("page", "")
	isAdmin := ctrl.PermissionService.IsUserAdmin(userID.(uint))

	response, status, err := ctrl.RetrospectiveService.List(userID.(uint), perPage, page, isAdmin)

	if err != nil {
		c.AbortWithStatusJSON(status, gin.H{"error": err.Error()})
	}

	c.JSON(status, response)
}

// Get Retrospective by id
func (ctrl RetrospectiveController) Get(c *gin.Context) {
	retroID := c.Param("retroID")
	userID, _ := c.Get("userID")

	if !ctrl.PermissionService.UserCanAccessRetro(retroID, userID.(uint)) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{})
		return
	}

	response, status, err := ctrl.RetrospectiveService.Get(retroID, true)
	if err != nil {
		c.AbortWithStatusJSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(status, response)
}

// GetEditLevels ...
func (ctrl RetrospectiveController) GetEditLevels(c *gin.Context) {
	retroID := c.Param("retroID")
	userID, _ := c.Get("userID")

	if !ctrl.PermissionService.UserCanAccessRetro(retroID, userID.(uint)) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{})
		return
	}

	response, status, err := ctrl.RetrospectiveService.GetEditLevels()
	if err != nil {
		c.AbortWithStatusJSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(status, response)
}

// GetTeamMembers ...
func (ctrl RetrospectiveController) GetTeamMembers(c *gin.Context) {
	userID, _ := c.Get("userID")
	retroID := c.Param("retroID")
	isAdmin := ctrl.PermissionService.IsUserAdmin(userID.(uint))

	if !ctrl.PermissionService.UserCanAccessRetro(retroID, userID.(uint)) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{})
		return
	}

	//ToDo: Match leaved_at with sprint dates instead of now
	members, status, err := ctrl.RetrospectiveService.GetTeamMembers(retroID, userID.(uint), isAdmin)
	if err != nil {
		c.AbortWithStatusJSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(status, members)
}

// GetLatestSprint returns the latest active/frozen sprint's data
func (ctrl RetrospectiveController) GetLatestSprint(c *gin.Context) {
	userID, _ := c.Get("userID")
	retroID := c.Param("retroID")

	if !ctrl.PermissionService.UserCanAccessRetro(retroID, userID.(uint)) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{})
		return
	}

	sprint, status, err := ctrl.RetrospectiveService.GetLatestSprint(retroID, userID.(uint))
	if err != nil {
		c.AbortWithStatusJSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(status, sprint)
}

// Create Retrospective
func (ctrl RetrospectiveController) Create(c *gin.Context) {
	userID, _ := c.Get("userID")
	var err error
	retrospectiveData := retrospectiveSerializers.RetrospectiveCreateSerializer{CreatedByID: userID.(uint)}
	if err = c.BindJSON(&retrospectiveData); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid request data"})
		return
	}

	retro, status, err := ctrl.RetrospectiveService.Create(userID.(uint), &retrospectiveData)
	if err != nil {
		c.AbortWithStatusJSON(status, gin.H{"error": err.Error()})
		return
	}

	ctrl.TrailService.Add(
		constants.CreatedRetrospective,
		constants.Retrospective,
		strconv.Itoa(int(retro.ID)),
		userID.(uint))

	c.JSON(status, retro)
}

// Update ...
func (ctrl RetrospectiveController) Update(c *gin.Context) {
	userID, _ := c.Get("userID")
	var err error
	retrospectiveData := retrospectiveSerializers.RetrospectiveUpdateSerializer{}
	if err = c.BindJSON(&retrospectiveData); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid request data"})
		return
	}

	retro, status, err := ctrl.RetrospectiveService.Update(userID.(uint), &retrospectiveData)
	if err != nil {
		c.AbortWithStatusJSON(status, gin.H{"error": err.Error()})
		return
	}

	ctrl.TrailService.Add(
		constants.UpdatedRetrospective,
		constants.Retrospective,
		strconv.Itoa(int(retro.ID)),
		userID.(uint))

	c.JSON(200, nil)
}
