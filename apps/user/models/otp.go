package models

import (
	"fmt"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/qor/admin"
)

// OTP ...
type OTP struct {
	gorm.Model
	Code   string `gorm:"type:varchar(16);not null"`
	User   User
	UserID uint `gorm:"not null"`
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
	return nil
}
