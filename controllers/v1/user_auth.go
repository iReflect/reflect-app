package v1

import (
	"fmt"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net/http"
	"strings"

	userModels "github.com/iReflect/reflect-app/apps/user/models"
	database "github.com/iReflect/reflect-app/db"
	"github.com/iReflect/reflect-app/libs/utils"
)

//UserAuthController ...
type UserAuthController struct {
}

//Add Routes
func (ctrl UserAuthController) Routes(r *gin.RouterGroup) {
	r.GET("/login", ctrl.Login)
	r.GET("/auth", ctrl.Auth)
	r.GET("/logout", ctrl.Logout)
}

// Login ...
func (ctrl UserAuthController) Login(c *gin.Context) {
	session := sessions.Default(c)
	state := session.Get("state")
	if state == nil {
		session.Set("state", utils.RandToken())
	}
	session.Save()
	c.JSON(http.StatusOK, state) // TODO Replace with oauth request
}

// Auth ...
func (ctrl UserAuthController) Auth(c *gin.Context) {
	session := sessions.Default(c)
	db, _ := database.GetFromContext(c)

	retrievedState := session.Get("state")
	actualState := c.Query("state")

	logrus.Info(fmt.Sprintf("State Expected:  %s, Actual: %s", retrievedState, actualState))

	if retrievedState != actualState {
		c.JSON(http.StatusUnauthorized, "Invalid request")
		return
	}

	// TODO replace with oauth logic
	email := c.Query("code")
	user := userModels.User{Email: email}
	if err := db.Preload("UserProfiles").Preload("Teams").First(&user).
		Error; err != nil {
		logrus.Error(err)
		c.JSON(http.StatusUnauthorized, "Invalid request")
		return
	}
	authToken := utils.RandToken()
	authSession := AuthSession{authToken, user}
	session.Set(authToken, authSession)
	session.Save()
	c.JSON(http.StatusOK, authSession)
}

type AuthSession struct {
	AuthToken string
	User      userModels.User
}

// Logout ...
func (ctrl UserAuthController) Logout(c *gin.Context) {
	tokenHeader := c.Request.Header.Get("Authorisation")
	tokenHeader = strings.Replace(tokenHeader, "Bearer ", "", -1)

	session := sessions.Default(c)
	session.Delete(tokenHeader)
	session.Save()
	c.JSON(http.StatusOK, "success")
}
