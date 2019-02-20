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
	//generate a hexa decimal value for "code" using time.Now().Unix().
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
