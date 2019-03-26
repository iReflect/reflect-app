package models

import (
	"fmt"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/qor/admin"

	"github.com/iReflect/reflect-app/constants"
)

// OTP ...
type OTP struct {
	Code     string `gorm:"type:varchar(16);primary_key"`
	ExpiryAt time.Time
	User     User
	UserID   uint `gorm:"unique"`
}

// RegisterOTPToAdmin ...
func RegisterOTPToAdmin(Admin *admin.Admin, config admin.Config) {
	otp := Admin.AddResource(&OTP{}, &config)

	otp.NewAttrs("-Code")
	otp.EditAttrs("-Code")
}

// BeforeCreate ...
func (otp *OTP) BeforeCreate(scope *gorm.Scope) error {
	//generate a hexadecimal value for "code" using time.Now().UnixNano().
	err := scope.SetColumn("Code", fmt.Sprintf("%X", time.Now().UnixNano()))
	if err != nil {
		return err
	}
	expiryAt := time.Unix(time.Now().Unix()+constants.OTPExpiryTime, 0)
	err = scope.SetColumn("ExpiryAt", expiryAt)
	if err != nil {
		return err
	}
	return nil
}

// GetReSendTime ...
func GetReSendTime(otp OTP) int {
	reSendTime := int(otp.ExpiryAt.Unix() - constants.OTPExpiryTime + constants.OTPReCreationTime - time.Now().Unix())
	if reSendTime < 0 {
		reSendTime = 0
	}
	return reSendTime
}
