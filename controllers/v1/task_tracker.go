package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
	taskTrackerServices "github.com/iReflect/reflect-app/apps/tasktracker/services"
)

//TaskTrackerController ...
type TaskTrackerController struct {
	TaskTrackerService taskTrackerServices.TaskTrackerService
}

//Routes for TaskTracker
func (ctrl TaskTrackerController) Routes(r *gin.RouterGroup) {
	r.GET("/config-list/", ctrl.ConfigList)
}

// ConfigList List task tracker config
func (ctrl TaskTrackerController) ConfigList(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"TaskProviders": ctrl.TaskTrackerService.ConfigList()})
}
