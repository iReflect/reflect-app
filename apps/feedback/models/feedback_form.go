package models

import (
	"github.com/jinzhu/gorm"
	"strconv"
	"github.com/qor/admin"
	"github.com/qor/qor"
	"github.com/qor/qor/resource"
	"github.com/sirupsen/logrus"

)

type FeedbackFormStatus int8

const (
	DraftFeedbackForm        FeedbackFormStatus = iota
	PublishedFeedbackForm
	ArchiveFeedbackForm
)

var FeedbackFormStatusValues = [...]string {
	"Draft",
	"Published",
	"Archived",
}

func (status FeedbackFormStatus) String() string {
	return FeedbackFormStatusValues[status]
}

// FeedbackForm represent template form for feedback
// TODO Add support for versioning
type FeedbackForm struct {
	gorm.Model
	Title       string `gorm:"type:varchar(255); not null"`
	Description string `gorm:"type:text;"`
	Status      FeedbackFormStatus   `gorm:"default:0; not null"`
	Archive     bool   `gorm:"default:false; not null"`
}

func RegisterFeedbackFormToAdmin(Admin *admin.Admin, config admin.Config) {
	feedbackForm := Admin.AddResource(&FeedbackForm{}, &config)
	statusMeta := getFeedbackFormStatusFieldMeta()
	feedbackForm.Meta(&statusMeta)
}

// getFeedbackFormStatusFieldMeta is the meta config for the feedback form status field
func getFeedbackFormStatusFieldMeta() admin.Meta {
	return admin.Meta{
		Name: "Status",
		Type: "select_one",
		Valuer: func(value interface{}, context *qor.Context) interface{} {
			feedbackForm := value.(*FeedbackForm)
			return strconv.Itoa(int(feedbackForm.Status))
		},
		Setter: func(resource interface{}, metaValue *resource.MetaValue, context *qor.Context) {
			feedbackForm := resource.(*FeedbackForm)
			value, err := strconv.Atoi(metaValue.Value.([]string)[0])
			if err != nil {
				logrus.Error("Cannot convert string to int")
				return
			}
			feedbackForm.Status = FeedbackFormStatus(value)
		},
		Collection: func(value interface{}, context *qor.Context) (results [][]string) {
			for index, value := range FeedbackFormStatusValues {
				results = append(results, []string{strconv.Itoa(index), value})
			}
			return
		},
		FormattedValuer: func(value interface{}, context *qor.Context) interface{} {
			feedbackForm := value.(*FeedbackForm)
			return feedbackForm.Status.String()
		},
	}
}
