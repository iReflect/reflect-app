package services

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/iReflect/reflect-app/libs/utils"
	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"
	"net/http"
	"strings"

	userModels "github.com/iReflect/reflect-app/apps/user/models"
	userSerializers "github.com/iReflect/reflect-app/apps/user/serializers"
)

//AuthenticationService ...
type AuthenticationService struct {
	DB *gorm.DB
}

// Login ...
func (service AuthenticationService) Login(c *gin.Context) string {
	session := sessions.Default(c)
	state := session.Get("state")
	if state == nil {
		state = utils.RandToken()
		session.Set("state", state)
	}
	session.Save()
	return state.(string)
}

// Authorize ...
func (service AuthenticationService) Authorize(c *gin.Context) (*userSerializers.UserAuthSerializer, error) {
	db := service.DB

	session := sessions.Default(c)
	retrievedState := session.Get("state")
	actualState := c.Query("state")

	logrus.Info(fmt.Sprintf("State Expected:  %s, Actual: %s", retrievedState, actualState))

	if retrievedState != actualState {
		return nil, errors.New("invalid state")
	}

	// TODO replace with oauth logic
	email := c.Query("code")
	user := userModels.User{Email: email}

	if err := db.First(&user).Error; err != nil {
		logrus.Info(fmt.Sprintf("User with email %s not found", email), err)
		return nil, err
	}

	authenticatedUser := new(userSerializers.UserAuthSerializer)

	db.Model(&user).Scan(&authenticatedUser)

	authenticatedUser.Token = utils.RandToken()
	session.Set("user", authenticatedUser.ID)
	session.Set("token", authenticatedUser.Token)
	session.Delete("state")

	session.Save()

	return authenticatedUser, nil
}

// Authenticate ...
func (service AuthenticationService) Authenticate(c *gin.Context) bool {
	db := service.DB

	token := getToken(c)
	session := sessions.Default(c)
	userID := session.Get("user")
	expectedToken := session.Get("token")
	if userID != nil && expectedToken != nil {
		user := userModels.User{}
		if err := db.First(&user, userID).Error; err != nil {
			logrus.Info(fmt.Sprintf("User with ID %s not found. Error: %s", userID, err))
			return false
		}
		if expectedToken == token && user.Active {
			c.Set("user", user)
			c.Set("userID", user.ID)
			return true
		}
	}

	return false
}

// Logout ...
func (service AuthenticationService) Logout(c *gin.Context) int {
	token := getToken(c)

	if service.Authenticate(c) {
		session := sessions.Default(c)
		session.Delete(token)
		session.Clear()
		session.Save()
		return http.StatusOK
	}
	return http.StatusUnauthorized
}

func getToken(c *gin.Context) string {
	token := c.Request.Header.Get("Authorization")
	return strings.Replace(token, "Bearer ", "", -1)

}
