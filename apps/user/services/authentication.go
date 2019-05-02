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
	errorCode string,
	err error) {

	var userData userSerializers.UserLogin
	err = c.BindJSON(&userData)
	if err != nil {
		logrus.Error(err)
		responseError := constants.APIErrorMessages[constants.InvalidRequestDataError]
		return nil, http.StatusBadRequest, responseError.Code, errors.New(responseError.Message)
	}

	gormDB := service.DB
	userResponse = new(userSerializers.UserAuthSerializer)

	err = gormDB.Model(&userModels.User{}).
		Where("email = ?", userData.Email).
		Scan(&userResponse).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return getInvalidEmailPasswordErrorResponse()
		}
		logrus.Error(err)
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
	return userResponse, http.StatusAccepted, "", nil
}

// Identify ...
func (service AuthenticationService) Identify(c *gin.Context) (
	reSendTime int,
	status int,
	errorCode string,
	err error) {

	var identifyData userSerializers.Identify
	err = c.BindJSON(&identifyData)
	if err != nil {
		logrus.Error(err)
		responseError := constants.APIErrorMessages[constants.InvalidRequestDataError]
		return 0, http.StatusBadRequest, responseError.Code, errors.New(responseError.Message)
	}
	gormDB := service.DB
	var userData userModels.User

	err = gormDB.Model(&userModels.User{}).Where("email = ?", identifyData.Email).Scan(&userData).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			responseError := constants.APIErrorMessages[constants.IReflectAccountNotFoundError]
			return 0, http.StatusBadRequest, responseError.Code, fmt.Errorf(responseError.Message+"%s", identifyData.Email)
		}
		logrus.Error(err)
		responseError := constants.APIErrorMessages[constants.SomethingWentWrong]
		return 0, http.StatusInternalServerError, responseError.Code, errors.New(responseError.Message)
	}

	var otp userModels.OTP
	err = gormDB.Model(&userModels.OTP{}).Where("user_id = ?", userData.ID).Scan(&otp).Error
	otpNotFound := gorm.IsRecordNotFoundError(err)
	if err != nil && !otpNotFound {
		logrus.Error(err)
		responseError := constants.APIErrorMessages[constants.SomethingWentWrong]
		return 0, http.StatusInternalServerError, responseError.Code, errors.New(responseError.Message)
	}

	// when we don't need to mail the OTP.
	if !identifyData.EmailOTP {
		if otpNotFound {
			responseError := constants.APIErrorMessages[constants.GeneratedOtpNotFoundError]
			return 0, http.StatusBadRequest, responseError.Code, errors.New(responseError.Message)
		}
		// if OTP is expired.
		if time.Now().Unix() > otp.ExpiryAt.Unix() {
			responseError := constants.APIErrorMessages[constants.GeneratedOtpExpiredError]
			return 0, http.StatusBadRequest, responseError.Code, errors.New(responseError.Message)
		}
		return otp.GetReSendTime(), http.StatusOK, "", nil
	}

	// if OTP exists then check its validity.
	if !otpNotFound {
		if otp.GetReSendTime() > 0 {
			responseError := constants.APIErrorMessages[constants.OtpReGeneratedError]
			return 0, http.StatusBadRequest, responseError.Code, errors.New(responseError.Message)
		}
		// deleting the old OTPs related to this email.
		err = gormDB.Delete(&otp).Error
		if err != nil {
			logrus.Error(err)
			responseError := constants.APIErrorMessages[constants.SomethingWentWrong]
			return 0, http.StatusInternalServerError, responseError.Code, errors.New(responseError.Message)
		}
	}

	newOTP := userModels.OTP{
		UserID: userData.ID,
	}

	// created new OTP for the user.
	err = gormDB.Create(&newOTP).Error
	if err != nil {
		logrus.Error(err)
		responseError := constants.APIErrorMessages[constants.SomethingWentWrong]
		return 0, http.StatusInternalServerError, responseError.Code, errors.New(responseError.Message)
	}

	// send this OTP via emaail.
	err = sendOTPAtEmail(identifyData.Email, newOTP.Code, userData.FirstName, userData.LastName)
	if err != nil {
		logrus.Error(err)
		responseError := constants.APIErrorMessages[constants.SomethingWentWrong]
		return 0, http.StatusInternalServerError, responseError.Code, errors.New(responseError.Message)
	}

	return newOTP.GetReSendTime(), http.StatusOK, "", nil
}

// Recover ...
func (service AuthenticationService) Recover(recoveryData userSerializers.Recover) (
	status int,
	errorCode string,
	err error) {

	gormDB := service.DB
	var userData userModels.User
	err = gormDB.Model(&userModels.User{}).Where("email = ?", recoveryData.Email).Scan(&userData).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			responseError := constants.APIErrorMessages[constants.IReflectAccountNotFoundError]
			return http.StatusBadRequest, responseError.Code, fmt.Errorf(responseError.Message+"%s", recoveryData.Email)
		}
		logrus.Error(err)
		responseError := constants.APIErrorMessages[constants.SomethingWentWrong]
		return http.StatusInternalServerError, responseError.Code, errors.New(responseError.Message)
	}
	var otp userModels.OTP
	err = gormDB.Model(&userModels.OTP{}).Where("user_id = ?", userData.ID).Scan(&otp).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			responseError := constants.APIErrorMessages[constants.GeneratedOtpNotFoundError]
			return http.StatusBadRequest, responseError.Code, errors.New(responseError.Message)
		}
		logrus.Error(err)
		responseError := constants.APIErrorMessages[constants.SomethingWentWrong]
		return http.StatusInternalServerError, responseError.Code, errors.New(responseError.Message)
	}

	if otp.Code != recoveryData.OTP {
		responseError := constants.APIErrorMessages[constants.InvalidOtpError]
		return http.StatusBadRequest, responseError.Code, errors.New(responseError.Message)
	}
	if otp.ExpiryAt.Unix() < time.Now().Unix() {
		responseError := constants.APIErrorMessages[constants.GeneratedOtpExpiredError]
		return http.StatusBadRequest, responseError.Code, errors.New(responseError.Message)
	}
	return http.StatusOK, "", nil
}

// UpdatePassword ...
func (service AuthenticationService) UpdatePassword(userPasswordData userSerializers.Recover) (
	status int,
	errorCode string,
	err error) {

	encryptedPassword := EncryptPassword(userPasswordData.Password)
	tx := service.DB.Begin()
	user := userModels.User{}

	err = tx.Model(&userModels.User{}).Where("email = ?", userPasswordData.Email).Update("password", encryptedPassword).Scan(&user).Error
	if err != nil {
		tx.Rollback()
		logrus.Error(err)
		responseError := constants.APIErrorMessages[constants.SomethingWentWrong]
		return http.StatusInternalServerError, responseError.Code, errors.New(responseError.Message)
	}

	err = tx.Where("user_id = ?", user.ID).Delete(&userModels.OTP{}).Error
	if err != nil {
		tx.Rollback()
		logrus.Error(err)
		responseError := constants.APIErrorMessages[constants.SomethingWentWrong]
		return http.StatusInternalServerError, responseError.Code, errors.New(responseError.Message)
	}

	err = tx.Commit().Error
	if err != nil {
		tx.Rollback()
		logrus.Error(err)
		responseError := constants.APIErrorMessages[constants.SomethingWentWrong]
		return http.StatusInternalServerError, responseError.Code, errors.New(responseError.Message)
	}
	return http.StatusOK, "", nil
}

func sendOTPAtEmail(email string, code string, firstName string, lastName string) error {
	message, err := parseTemplate("apps/user/views/mail.html", map[string]interface{}{"firstName": firstName, "lastName": lastName, "code": code})
	if err != nil {
		logrus.Error(err)
		return err
	}
	// get email configrations from environment variables.
	emailConfig := config.GetConfig().Email
	subject := fmt.Sprintf("Subject: %s\n", constants.OTPEmailSubject)
	emailFrom := fmt.Sprintf("From: %s\n", emailConfig.EmailFrom)
	to := fmt.Sprintf("To: %s\n", email)
	// TODO: serching for way to send both type of bodies i.e html and text mail.
	body := []byte(fmt.Sprintf("%s%s%s%s%s", subject, emailFrom, to, constants.EmailMIME, message))

	// Set up authentication information.
	auth := smtp.PlainAuth(
		"",
		emailConfig.Username,
		emailConfig.Password,
		emailConfig.Host,
	)

	// Connect to the server, authenticate, set the sender and recipient,
	// and send the email all in one step.
	err = smtp.SendMail(
		fmt.Sprintf("%s:%s", emailConfig.Host, emailConfig.Port),
		auth,
		emailConfig.EmailFrom,
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
	errorCode string,
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

	return userResponse, http.StatusOK, "", nil
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
	errorCode string,
	err error) {
	responseError := constants.APIErrorMessages[constants.InternalServerError]
	return nil, http.StatusInternalServerError, responseError.Code, errors.New(responseError.Message)
}

// getNotFoundErrorResponse ...
func getNotFoundErrorResponse() (authenticatedUser *userSerializers.UserAuthSerializer,
	status int,
	errorCode string,
	err error) {
	responseError := constants.APIErrorMessages[constants.UserNotFoundError]
	return nil, http.StatusNotFound, responseError.Code, errors.New(responseError.Message)
}

// getInvalidEmailPasswordErrorResponse ...
func getInvalidEmailPasswordErrorResponse() (authenticatedUser *userSerializers.UserAuthSerializer,
	status int,
	errorCode string,
	err error) {
	responseError := constants.APIErrorMessages[constants.InvalidEmailOrPassword]
	return nil, http.StatusNotFound, responseError.Code, errors.New(responseError.Message)
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
