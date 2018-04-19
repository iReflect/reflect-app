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
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/plus/v1"
	"net/http"
	"os"
)

var googleOAuthConf *oauth2.Config

func init() {
	var err error
	googleOAuthConf, err = getGoogleOAuthConf()
	if err != nil {
		os.Exit(1)
	}

}

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
		"LoginURL": googleOAuthConf.AuthCodeURL(state.(string)),
	}
}

// Authorize ...
func (service AuthenticationService) Authorize(c *gin.Context) (
	userResponse *userSerializers.UserAuthSerializer,
	status int,
	err error) {
	db := service.DB

	oAuthContext := context.TODO()

	session := sessions.Default(c)
	retrievedState := session.Get("state")
	actualState := c.Query("state")

	resetSession(session)

	if retrievedState != actualState {
		logrus.Error(fmt.Sprintf("State Expected:  %s, Actual: %s", retrievedState, actualState))
		return getNotFoundErrorResponse()
	}

	tok, err := googleOAuthConf.Exchange(oAuthContext, c.Query("code"))
	if err != nil {
		logrus.Error("Error occurred while exchanging code with token, Error:", err)
		return getNotFoundErrorResponse()
	}

	client := googleOAuthConf.Client(oAuthContext, tok)

	plusService, err := plus.New(client)
	if err != nil {
		logrus.Error("Error occurred while creating google plus service, Error:", err)
		return getInternalErrorResponse()
	}

	googleUser, err := plusService.People.Get("me").Do()
	if err != nil {
		logrus.Error("Error occurred while getting information from google, Error:", err)
		return getInternalErrorResponse()
	}
	userEmail := getAccountEmail(googleUser)
	user := userModels.User{}
	if err := db.
		Where("users.deleted_at IS NULL").
		Where("email = ?", userEmail).
		First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			logrus.Info(fmt.Sprintf("User with email %s not found", userEmail))
			return getNotFoundErrorResponse()
		}
		logrus.Error("Error occurred while getting user from DB, Error:", err)
		return getInternalErrorResponse()
	}

	userResponse = new(userSerializers.UserAuthSerializer)

	db.Model(&user).
		Where("users.deleted_at IS NULL").
		Scan(&userResponse)

	userResponse.Token = utils.RandToken()
	session.Set("user", userResponse.ID)
	session.Set("token", userResponse.Token)
	session.Save()

	logrus.Info(fmt.Sprintf("Logged in user %s", userResponse.Email))

	return userResponse, http.StatusOK, nil
}

// AuthenticateSession ...
func (service AuthenticationService) AuthenticateSession(c *gin.Context) bool {
	db := service.DB

	session := sessions.Default(c)
	userID := session.Get("user")
	if userID != nil {
		authenticatedUser := userModels.User{}
		if err := db.
			Where("users.deleted_at IS NULL").
			Where("active = true").
			First(&authenticatedUser, userID).Error; err != nil {
			logrus.Error(fmt.Sprintf("User with ID %s not found. Error: %s", userID, err))
			return false
		}
		logrus.Info(fmt.Sprintf("Authenticated user %s", authenticatedUser.Email))
		c.Set("user", authenticatedUser)
		c.Set("userID", authenticatedUser.ID)
		return true
	}

	return false
}

// Logout ...
func (service AuthenticationService) Logout(c *gin.Context) int {

	if service.AuthenticateSession(c) {
		session := sessions.Default(c)
		currentUser, _ := c.Get("user")
		user := currentUser.(userModels.User)
		resetSession(session)
		logrus.Info(fmt.Sprintf("Logged out user %s", user.Email))

		return http.StatusOK
	}
	return http.StatusUnauthorized
}

// getAccountEmail ...
func getAccountEmail(person *plus.Person) string {
	personEmails := person.Emails

	for _, personEmail := range personEmails {
		if personEmail.Type == "account" {
			return personEmail.Value
		}
	}
	logrus.Error("No account email found")
	return ""
}

// resetSession ...
func resetSession(session sessions.Session) {
	session.Set("user", nil)
	session.Set("token", nil)
	session.Set("state", nil)
	session.Clear()
	session.Save()
}

// getUnauthorizedErrorResponse ...
func getInternalErrorResponse() (authenticatedUser *userSerializers.UserAuthSerializer,
	status int,
	err error) {
	return nil, http.StatusInternalServerError, errors.New("internal server error")
}

// getNotFoundErrorResponse ...
func getNotFoundErrorResponse() (authenticatedUser *userSerializers.UserAuthSerializer,
	status int,
	err error) {
	return nil, http.StatusNotFound, errors.New("user not found")
}

// getGoogleOAuthConf ...
func getGoogleOAuthConf() (*oauth2.Config, error) {
	credentials, err := google.FindDefaultCredentials(context.TODO())

	if err != nil {
		logrus.Error("error loading google creds, Error:", err)
		return nil, err
	}

	oauthConfig, err := google.ConfigFromJSON(credentials.JSON, plus.PlusMeScope,
		plus.UserinfoEmailScope, plus.UserinfoProfileScope)

	if err != nil {
		logrus.Error("error loading google creds, Error", err)
		return nil, err
	}
	oauthConfig.Endpoint = google.Endpoint

	return oauthConfig, nil
}
