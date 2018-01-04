package services

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	userModels "github.com/iReflect/reflect-app/apps/user/models"
	userSerializers "github.com/iReflect/reflect-app/apps/user/serializers"
	"github.com/iReflect/reflect-app/libs/utils"
	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"
	"net/http"
	"net/url"
)

//AuthenticationService ...
type AuthenticationService struct {
	DB *gorm.DB
}

// Login ...
func (service AuthenticationService) Login(c *gin.Context) map[string]string {
	session := sessions.Default(c)
	state := session.Get("state")
	if state == nil {
		state = utils.RandToken()
		session.Set("state", state)
	}
	session.Save()
	return map[string]string{
		"State": state.(string),
	}
}

// Authorize ...
func (service AuthenticationService) Authorize(c *gin.Context) (
	authenticatedUser *userSerializers.UserAuthSerializer,
	err error) {
	db := service.DB

	session := sessions.Default(c)
	retrievedState := session.Get("state")
	actualState := c.Query("state")

	if retrievedState != actualState {
		logrus.Info(fmt.Sprintf("State Expected:  %s, Actual: %s", retrievedState, actualState))
		return nil, errors.New("invalid state")
	}

	// TODO replace with oauth logic
	email, _ := url.QueryUnescape(c.Query("code"))
	user := userModels.User{}

	if err := db.Where("email = ?", email).First(&user).Error; err != nil {
		logrus.Info(fmt.Sprintf("User with email %s not found", email), err)
		session.Set("user", nil)
		session.Set("token", nil)
		session.Save()
		return nil, err
	}

	authenticatedUser = new(userSerializers.UserAuthSerializer)

	db.Model(&user).Scan(&authenticatedUser)

	authenticatedUser.Token = utils.RandToken()
	session.Set("user", authenticatedUser.ID)
	session.Set("token", authenticatedUser.Token)

	session.Save()
	logrus.Info(fmt.Sprintf("Logged in user %s", authenticatedUser.Email))

	return authenticatedUser, nil
}

// AuthenticateSession ...
func (service AuthenticationService) AuthenticateSession(c *gin.Context) bool {
	db := service.DB

	session := sessions.Default(c)
	userID := session.Get("user")
	if userID != nil {
		authenticatedUser := userModels.User{}
		if err := db.First(&authenticatedUser, userID).Error; err != nil {
			logrus.Info(fmt.Sprintf("User with ID %s not found. Error: %s", userID, err))
			return false
		}
		if authenticatedUser.Active {
			logrus.Info(fmt.Sprintf("Authenticated user %s", authenticatedUser.Email))
			c.Set("user", authenticatedUser)
			c.Set("userID", authenticatedUser.ID)
			return true
		}
	}

	return false
}

// Logout ...
func (service AuthenticationService) Logout(c *gin.Context) int {

	if service.AuthenticateSession(c) {
		session := sessions.Default(c)
		currentUser, _ := c.Get("user")
		user := currentUser.(userModels.User)
		session.Set("user", nil)
		session.Set("token", nil)
		session.Set("state", nil)
		logrus.Info(fmt.Sprintf("Logged out user %s", user.Email))

		session.Clear()
		session.Save()
		return http.StatusOK
	}
	return http.StatusUnauthorized
}
