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
		responseError := constants.APIErrorMessages[constants.TaskTrackerNameIsMustError]
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": responseError.Message, "code": responseError.Code})
		return
	}
	team, exists := c.GetQuery("teamID")
	if !exists {
		responseError := constants.APIErrorMessages[constants.TeamIDIsMustError]
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": responseError.Message, "code": responseError.Code})
		return
	}
	timeTrackers, status, errorCode, err := ctrl.TaskTrackerService.SupportedTimeTrackersList(taskTracker, team)
	if err != nil {
		c.AbortWithStatusJSON(status, gin.H{"error": err.Error(), "code": errorCode})
		return
	}
	c.JSON(status, timeTrackers)
}
