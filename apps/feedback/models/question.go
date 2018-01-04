package models

import (
	"github.com/jinzhu/gorm"
	"github.com/qor/admin"
	"github.com/qor/qor"
	"github.com/qor/qor/resource"

	"github.com/iReflect/reflect-app/db/models/fields"
	"github.com/iReflect/reflect-app/libs/utils"
	"strconv"
	"github.com/sirupsen/logrus"
	"errors"
)


type QuestionType int8

const (
	MultiChoiceType QuestionType = iota
	GradingType
	BooleanType
)

var QuestionTypeValues = [...]string {
	"Multi Choice",
	"Grade",
	"Boolean",
}

func (questionType QuestionType) String() string {
	return QuestionTypeValues[questionType]
}

// Question represent the questions asked for a skill
// TODO Add support for versioning and making it non-editable
type Question struct {
	gorm.Model
	Text    string `gorm:"type:text; not null"`
	Type    QuestionType   `gorm:"default:0; not null)"`
	Skill   Skill
	SkillID uint         `gorm:"not null"`
	Options fields.JSONB `gorm:"type:jsonb; not null; default:'{}'::jsonb"`
	Weight  int          `gorm:"default:1; not null"`
}

func (question Question) GetOptions() map[string]interface{} {
	return utils.ByteToMap(question.Options)
}

// ValidateQuestionResponse validates the question response (default also), against the question options
func (question *Question) ValidateQuestionResponse(questionResponse string) bool {

	questionResponseList := GetQuestionResponseList(questionResponse)

	// If the question response is not an empty string but the question response list is,
	// there is some error in the response format
	if questionResponse != "" && len(questionResponseList) == 0 {
		return false
	}

	// Check even if the response is valid based on question type
	if question.Type == BooleanType || question.Type == GradingType {
		if len(questionResponseList) > 1 {
			return false
		}
	}
	questionOptions := question.GetOptions()
	validValues := map[float64]float64{}
	for _, val := range questionOptions["values"].([]interface{}) {
		responseID := val.(map[string]interface{})["id"].(float64)
		validValues[responseID] = responseID
	}
	for _, response := range questionResponseList {
		if response != "" {
			value, _ := strconv.ParseFloat(response, 64)
			_, isValid := validValues[value]
			if !isValid {return isValid}
		}
	}
	return true
}

func (question *Question) BeforeSave(db *gorm.DB) (err error) {

	// Check if default question response is valid
	defaultOptions, exists := question.GetOptions()["defaultValue"]
	if exists && defaultOptions != "" {
		if isValid := question.ValidateQuestionResponse(defaultOptions.(string)); !isValid {
			err = errors.New("default value can only be from valid values")
		}
	}
	return
}

func RegisterQuestionToAdmin(Admin *admin.Admin, config admin.Config) {
	question := Admin.AddResource(&Question{}, &config)
	optionsMeta := getOptionsFieldMeta()
	typesMeta := getTypeFieldMeta()
	question.Meta(&optionsMeta)
	question.Meta(&typesMeta)

}

func SetQuestionRelatedFieldMeta(res *admin.Resource) {
	optionsMeta := getOptionsFieldMeta()
	typesMeta := getTypeFieldMeta()
	res.Meta(&optionsMeta)
	res.Meta(&typesMeta)
}

// getOptionsFieldMeta is the meta config for the question's options field
func getOptionsFieldMeta() admin.Meta {
	return admin.Meta{
		Name: "Options",
		Type: "text",
		Valuer: func(value interface{}, context *qor.Context) interface{} {
			question := value.(*Question)
			return string(question.Options)
		},
		Setter: func(resource interface{}, metaValue *resource.MetaValue, context *qor.Context) {
			question := resource.(*Question)
			value := metaValue.Value.([]string)[0]
			question.Options = fields.JSONB(value)
		}}

}

// getTypeFieldMeta is the meta config for the question's type field
func getTypeFieldMeta() admin.Meta {
	return admin.Meta{
		Name: "Type",
		Type: "select_one",
		Valuer: func(value interface{}, context *qor.Context) interface{} {
			question := value.(*Question)
			return strconv.Itoa(int(question.Type))
		},
		Setter: func(resource interface{}, metaValue *resource.MetaValue, context *qor.Context) {
			question := resource.(*Question)
			value, err := strconv.Atoi(metaValue.Value.([]string)[0])
			if err != nil {
				logrus.Error("Cannot convert string to int")
				return
			}
			question.Type = QuestionType(value)
		},
		Collection: func(value interface{}, context *qor.Context) (results [][]string) {
			for index, value := range QuestionTypeValues {
				results = append(results, []string{strconv.Itoa(index), value})
			}
			return
		},
		FormattedValuer: func(value interface{}, context *qor.Context) interface{} {
			question := value.(*Question)
			return question.Type.String()
		},
	}
}
