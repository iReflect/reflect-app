package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
	userServices "github.com/iReflect/reflect-app/apps/user/services"
)

//TeamController ...
type TeamController struct {
	TeamService userServices.TeamService
}

// Routes for Team
func (ctrl TeamController) Routes(r *gin.RouterGroup) {
	r.GET("/:teamID/members/", ctrl.GetMembers)
	r.GET("/", ctrl.GetTeams)
}

// GetTeams ...
func (ctrl TeamController) GetTeams(c *gin.Context) {
	userID, _ := c.Get("userID")
	teams, err := ctrl.TeamService.UserTeamList(userID.(uint), true)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "Could not get teams", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, teams)
}

// GetMembers ...
func (ctrl TeamController) GetMembers(c *gin.Context) {
	id := c.Param("teamID")
	all := c.DefaultQuery("all", "false")
	userID, _ := c.Get("userID")

	members, err := ctrl.TeamService.MemberList(id, userID.(uint), all != "true")

	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "Could not get members", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, members)
}
