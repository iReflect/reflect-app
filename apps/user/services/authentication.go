package services

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"net/smtp"
	"os"
	"reflect"
	"time"

	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/pbkdf2"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	googleOAuthAPI "google.golang.org/api/oauth2/v2"

	userModels "github.com/iReflect/reflect-app/apps/user/models"
	userSerializers "github.com/iReflect/reflect-app/apps/user/serializers"
	"github.com/iReflect/reflect-app/config"
	"github.com/iReflect/reflect-app/constants"
	"github.com/iReflect/reflect-app/libs/utils"
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

// BasicLogin ...
func (service AuthenticationService) BasicLogin(c *gin.Context) (
	userResponse *userSerializers.UserAuthSerializer,
	status int,
	err error) {

	var userData userSerializers.UserLogin
	err = c.BindJSON(&userData)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid request data"})
		return
	}

	gormDB := service.DB
	userResponse = new(userSerializers.UserAuthSerializer)

	err = gormDB.Model(&userModels.User{}).
		Where("email = ?", userData.Email).
		Scan(&userResponse).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return getInvalidEmailPasswordErrorResponse()
		}
		return getInternalErrorResponse()
	}
	// here we encrypt password before comparing it with stored password because we store passwords after encryption.
	encryptedPassword := EncryptPassword(userData.Password)
	if !reflect.DeepEqual(encryptedPassword, userResponse.Password) || userResponse.Password == nil {
		return getInvalidEmailPasswordErrorResponse()
	}

	session := sessions.Default(c)
	userResponse.Token = utils.RandToken()
	setSession(session, userResponse)
	return userResponse, http.StatusAccepted, nil
}

// Identify ...
func (service AuthenticationService) Identify(c *gin.Context) (
	status int,
	err error) {

	var identifyData userSerializers.Identify
	err = c.BindJSON(&identifyData)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid request data"})
		return
	}
	gormDB := service.DB
	var userData userModels.User
	err = gormDB.Model(&userModels.User{}).Where("email = ?", identifyData.Email).Scan(&userData).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return http.StatusBadRequest, fmt.Errorf("We couldn't find a iReflect account associated with %s", identifyData.Email)
		}
		return http.StatusInternalServerError, err
	}
	// when we don't need OTP to the email.
	if !identifyData.EmailOTP {
		return http.StatusOK, nil
	}
	var otp userModels.OTP
	err = gormDB.Model(&userModels.OTP{}).Where("user_id = ?", userData.ID).Scan(&otp).Error
	// when there is some internal DB error.
	if err != nil && !gorm.IsRecordNotFoundError(err) {
		return http.StatusInternalServerError, err
	}
	// if OTP exists then check its validity.
	if !gorm.IsRecordNotFoundError(err) {
		if otp.ExpiryAt.Unix()-constants.OTPExpiryTime+constants.OTPReCreationTime > time.Now().Unix() {
			return http.StatusBadRequest, errors.New("You just generated a OTP. Please try again after sometime")
		}
		err = gormDB.Delete(&otp).Error
		if err != nil {
			return http.StatusInternalServerError, err
		}
	}
	newOTP := userModels.OTP{
		UserID: userData.ID,
	}
	err = gormDB.Create(&newOTP).Error
	if err != nil {
		return http.StatusInternalServerError, err
	}
	err = sendOTPAtEmail(identifyData.Email, newOTP.Code, userData.FirstName, userData.LastName)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	return http.StatusOK, nil
}

func sendOTPAtEmail(email string, code string, firstName string, lastName string) error {
	message, _ := parseTemplate("apps/user/views/mail.html", map[string]interface{}{"firstName": firstName, "lastName": lastName, "code": code})

	to := fmt.Sprintf("To: %s\n", email)
	// TODO: serching for way to send both type of bodies i.e html and text mail.
	body := []byte(constants.OTPEmailSubject + constants.EmailFrom + to + constants.EmailMIME + "\n" + message)
	// get email configrations from environment variables.
	emailConfig := config.GetConfig().Email
	// Set up authentication information.
	auth := smtp.PlainAuth(
		"",
		emailConfig.Username,
		emailConfig.Password,
		emailConfig.Host,
	)

	// Connect to the server, authenticate, set the sender and recipient,
	// and send the email all in one step.
	err := smtp.SendMail(
		fmt.Sprintf("%s:%s", emailConfig.Host, emailConfig.Port),
		auth,
		constants.IReflectEmail,
		[]string{email},
		body,
	)
	if err != nil {
		logrus.Error(err)
		return err
	}
	return nil
}

func parseTemplate(fileName string, data interface{}) (string, error) {
	t, err := template.ParseFiles(fileName)
	if err != nil {
		return "", err
	}
	buffer := new(bytes.Buffer)
	if err = t.Execute(buffer, data); err != nil {
		return "", err
	}
	return buffer.String(), nil
}

// EncryptPassword ...
func EncryptPassword(password string) []byte {
	return pbkdf2.Key([]byte(password), []byte(constants.PasswordSalt), constants.IterationCount, constants.KeyLength, sha256.New)
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

	oauthService, err := googleOAuthAPI.New(client)
	if err != nil {
		logrus.Error("Error occurred while creating google oauth service, Error:", err)
		return getInternalErrorResponse()
	}

	googleUser, err := oauthService.Userinfo.Get().Do()
	if err != nil {
		logrus.Error("Error occurred while getting information from google, Error:", err)
		return getInternalErrorResponse()
	}
	userEmail := googleUser.Email
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
	setSession(session, userResponse)
	logrus.Info(fmt.Sprintf("Logged in user %s", userResponse.Email))

	return userResponse, http.StatusOK, nil
}

// AuthenticateSession ...
func (service AuthenticationService) AuthenticateSession(c *gin.Context) (int, error) {
	db := service.DB

	session := sessions.Default(c)
	userID := session.Get("user")
	if userID != nil {
		authenticatedUser := userModels.User{}
		err := db.Where("users.deleted_at IS NULL").Where("active = true").First(&authenticatedUser, userID).Error
		if err != nil {
			if gorm.IsRecordNotFoundError(err) {
				logrus.Error(fmt.Sprintf("User with ID %s not found. Error: %s", userID, err))
				return http.StatusUnauthorized, fmt.Errorf("User with ID %s not found", userID)
			}
			logrus.Error(err)
			return http.StatusInternalServerError, err
		}
		logrus.Info(fmt.Sprintf("Authenticated user %s", authenticatedUser.Email))
		c.Set("user", authenticatedUser)
		c.Set("userID", authenticatedUser.ID)
		return http.StatusOK, nil
	}

	return http.StatusUnauthorized, fmt.Errorf("User with ID %s not found", userID)
}

// Logout ...
func (service AuthenticationService) Logout(c *gin.Context) int {

	status, err := service.AuthenticateSession(c)
	if err != nil {
		return status
	}
	session := sessions.Default(c)
	currentUser, _ := c.Get("user")
	user := currentUser.(userModels.User)
	resetSession(session)
	logrus.Info(fmt.Sprintf("Logged out user %s", user.Email))

	return http.StatusOK

}

// setSession ...
func setSession(session sessions.Session, userResponse *userSerializers.UserAuthSerializer) {
	session.Set("user", userResponse.ID)
	session.Set("token", userResponse.Token)
	session.Save()
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

// getInvalidEmailPasswordErrorResponse ...
func getInvalidEmailPasswordErrorResponse() (authenticatedUser *userSerializers.UserAuthSerializer,
	status int,
	err error) {
	return nil, http.StatusNotFound, errors.New(constants.InvalidEmailOrPassword)
}

// getGoogleOAuthConf ...
func getGoogleOAuthConf() (*oauth2.Config, error) {
	credentials, err := google.FindDefaultCredentials(context.TODO())

	if err != nil {
		logrus.Error("error loading google creds, Error:", err)
		return nil, err
	}

	oauthConfig, err := google.ConfigFromJSON(credentials.JSON, googleOAuthAPI.UserinfoEmailScope, googleOAuthAPI.UserinfoProfileScope)
	if err != nil {
		logrus.Error("error loading google creds, Error", err)
		return nil, err
	}
	oauthConfig.Endpoint = google.Endpoint

	return oauthConfig, nil
}
