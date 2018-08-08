package v1

import (
	"strconv"

	"github.com/gin-gonic/gin"
	trailService "github.com/iReflect/reflect-app/apps/retrospective/services"
)

// TrailController ...
type TrailController struct {
	TrailService trailService.TrailService
}

// Routes for Trail
func (ctrl TrailController) Routes(r *gin.RouterGroup) {
	r.GET("/:sprintId/process_history", ctrl.GetTrail)
}

// GetTrail is method to get the all trails related to a particular sprint
func (ctrl TrailController) GetTrail(c *gin.Context) {
	sprintID, _ := strconv.Atoi(c.Param("sprintId"))

	trails, status, err := ctrl.TrailService.GetTrails(uint(sprintID))

	if err != nil {
		return
	}

	c.JSON(status, trails)
}
