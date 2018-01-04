package models

import (
	"time"

	"github.com/jinzhu/gorm"

	userModels "github.com/iReflect/reflect-app/apps/user/models"
	"strconv"
	"github.com/qor/admin"
	"github.com/qor/qor"
	"github.com/qor/qor/resource"
	"github.com/sirupsen/logrus"
)

var FeedbackStatusValues = [...]string {
	"New",
	"In Progress",
	"Submitted",
}

type FeedbackStatus int8

func (status FeedbackStatus) String() string {
	return FeedbackStatusValues[status]
}

const (
	NewFeedback        FeedbackStatus = iota
	InProgressFeedback
	SubmittedFeedback
)


// Feedback represent a submitted/in-progress feedback form by a user
type Feedback struct {
	gorm.Model
	FeedbackForm     FeedbackForm
	Title            string `gorm:"type:varchar(255); not null"`
	FeedbackFormID   uint   `gorm:"not null"`
	ForUserProfile   userModels.UserProfile
	ForUserProfileID uint
	ByUserProfile    userModels.UserProfile
	ByUserProfileID  uint `gorm:"not null"`
	Team             userModels.Team
	TeamID           uint `gorm:"not null"`
	Status           FeedbackStatus `gorm:"default:0; not null"`
	SubmittedAt      *time.Time
	DurationStart    time.Time `gorm:"not null"`
	DurationEnd      time.Time `gorm:"not null"`
	ExpireAt         time.Time `gorm:"not null"`
}


func RegisterFeedbackToAdmin(Admin *admin.Admin, config admin.Config) {
	feedback := Admin.AddResource(&Feedback{}, &config)
	statusMeta := getStatusFieldMeta()
	feedback.Meta(&statusMeta)
}

// getStatusFieldMeta is the meta config for the feedback status field
func getStatusFieldMeta() admin.Meta {
	return admin.Meta{
		Name: "Status",
		Type: "select_one",
		Valuer: func(value interface{}, context *qor.Context) interface{} {
			feedback := value.(*Feedback)
			return strconv.Itoa(int(feedback.Status))
		},
		Setter: func(resource interface{}, metaValue *resource.MetaValue, context *qor.Context) {
			feedback := resource.(*Feedback)
			value, err := strconv.Atoi(metaValue.Value.([]string)[0])
			if err != nil {
				logrus.Error("Cannot convert string to int")
				return
			}
			feedback.Status = FeedbackStatus(value)
		},
		Collection: func(value interface{}, context *qor.Context) (results [][]string) {
			for index, value := range FeedbackStatusValues {
				results = append(results, []string{strconv.Itoa(index), value})
			}
			return
		},
		FormattedValuer: func(value interface{}, context *qor.Context) interface{} {
			feedback := value.(*Feedback)
			return feedback.Status.String()
		},
	}
}
