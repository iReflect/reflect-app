package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"

	taskTrackerServices "github.com/iReflect/reflect-app/apps/tasktracker/services"
	"github.com/iReflect/reflect-app/constants"
)

//TaskTrackerController ...
type TaskTrackerController struct {
	TaskTrackerService taskTrackerServices.TaskTrackerService
}

//Routes for TaskTracker
func (ctrl TaskTrackerController) Routes(r *gin.RouterGroup) {
	r.GET("/config-list/", ctrl.ConfigList)
	r.GET("/supported-time-providers/", ctrl.SupportedTimeTrackersList)
}

// ConfigList List task tracker config
func (ctrl TaskTrackerController) ConfigList(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"TaskProviders": ctrl.TaskTrackerService.ConfigList()})
}

// SupportedTimeTrackersList ...
func (ctrl TaskTrackerController) SupportedTimeTrackersList(c *gin.Context) {
	taskTracker, exists := c.GetQuery("taskTrackerName")
	if !exists {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": constants.TaskTrackerNameIsMustError})
		return
	}
	team, exists := c.GetQuery("teamID")
	if !exists {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": constants.TeamIDIsMustError})
		return
	}
	timeTrackerList, err := ctrl.TaskTrackerService.SupportedTimeTrackersList(taskTracker, team)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"timeProviderList": timeTrackerList})
}
