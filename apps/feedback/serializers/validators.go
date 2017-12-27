package serializers

import (
	"reflect"
	"regexp"
	"time"

	mapset "github.com/deckarep/golang-set"

	"github.com/jinzhu/gorm"

	feedbackModels "github.com/iReflect/reflect-app/apps/feedback/models"
	validator "gopkg.in/go-playground/validator.v8"
)

// Regex to test the response format of a question
const questionResponseRegexString = "^[0-9]+(,[0-9]+)*$"

var questionResponseRegex = regexp.MustCompile(questionResponseRegexString)

// IsValidQuestionResponse validates the "Response" value of QuestionResponseSerializer
func IsValidQuestionResponse(
	v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value,
	field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string,
) bool {
	feedbackStatus := topStruct.Interface().(*FeedbackResponseSerializer).Status
	questionResponse := field.String()
	// Response can be an empty string only if the feedback is not getting submitted
	if feedbackStatus != 2 && questionResponse == "" {
		return true
	}
	return questionResponseRegex.MatchString(questionResponse)
}

// IsValidSubmittedAt validates the "SubmittedAt" value of FeedbackResponseSerializer
func IsValidSubmittedAt(
	v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value,
	field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string,
) bool {
	feedbackStatus := currentStructOrField.Interface().(*FeedbackResponseSerializer).Status
	if feedbackStatus == 2 {
		// Check if the submitted at is in correct format
		_, err := time.Parse(time.RFC3339, field.String())
		if err != nil {
			return false
		}
	}
	return true
}

// IsValidSaveAndSubmit validates the "SaveAndSubmit" value of FeedbackResponseSerializer
func IsValidSaveAndSubmit(
	v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value,
	field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string,
) bool {
	saveAndSubmit := field.Bool()
	feedbackStatus := currentStructOrField.Interface().(*FeedbackResponseSerializer).Status
	if feedbackStatus == 2 {
		// If status is 2 (submitted) then saveAndSubmit should be true
		return saveAndSubmit
	}
	return true
}

// IsAllQuestionPresent validates the "Data" value of FeedbackResponseSerializer
func IsAllQuestionPresent(db *gorm.DB) validator.Func {
	return func(
		v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value,
		field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string,
	) bool {
		var expectedResponseIDList, actualResponseIDList []interface{}
		feedbackResponseData := field.Interface().(FeedbackResponseData)
		feedbackID := currentStructOrField.Interface().(*FeedbackResponseSerializer).FeedbackID
		if err := db.Model(feedbackModels.QuestionResponse{}).
			Where("feedback_id = ?", feedbackID).
			Pluck("id", &expectedResponseIDList).Error; err != nil {
			return false
		}
		for _, categoryData := range feedbackResponseData {
			for _, skillData := range categoryData {
				for questionResponseID := range skillData {
					actualResponseIDList = append(actualResponseIDList, questionResponseID)
				}
			}
		}
		expectedResponseIDSet := mapset.NewSetFromSlice(expectedResponseIDList)
		actualResponseIDSet := mapset.NewSetFromSlice(actualResponseIDList)
		return expectedResponseIDSet.Equal(actualResponseIDSet)
	}
}
