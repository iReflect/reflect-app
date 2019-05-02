package v1

import (
	"github.com/gin-gonic/gin"
	retrospectiveService "github.com/iReflect/reflect-app/apps/retrospective/services"
	userServices "github.com/iReflect/reflect-app/apps/user/services"
)

//TeamController ...
type TeamController struct {
	TeamService       userServices.TeamService
	PermissionService retrospectiveService.PermissionService
}

// Routes for Team
func (ctrl TeamController) Routes(r *gin.RouterGroup) {
	r.GET("/:teamID/members/", ctrl.GetMembers)
	r.GET("/", ctrl.GetTeams)
}

// GetTeams ...
func (ctrl TeamController) GetTeams(c *gin.Context) {
	userID, _ := c.Get("userID")
	teams, status, errorCode, err := ctrl.TeamService.UserTeamList(userID.(uint), true)

	if err != nil {
		c.AbortWithStatusJSON(status, gin.H{"error": err.Error(), "code": errorCode})
		return
	}

	c.JSON(status, teams)
}

// GetMembers ...
func (ctrl TeamController) GetMembers(c *gin.Context) {
	id := c.Param("teamID")
	all := c.DefaultQuery("all", "false")
	userID, _ := c.Get("userID")
	isAdmin := ctrl.PermissionService.IsUserAdmin(userID.(uint))

	members, status, errorCode, err := ctrl.TeamService.MemberList(id, userID.(uint), all != "true", isAdmin)

	if err != nil {
		c.AbortWithStatusJSON(status, gin.H{"error": err.Error(), "code": errorCode})
		return
	}

	c.JSON(status, members)
}
