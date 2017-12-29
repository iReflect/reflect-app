package validators

import (
	"reflect"
	"time"

	"github.com/deckarep/golang-set"

	"github.com/jinzhu/gorm"
	"github.com/gin-gonic/gin/binding"

	feedbackModels "github.com/iReflect/reflect-app/apps/feedback/models"
	feedbackSerializers "github.com/iReflect/reflect-app/apps/feedback/serializers"
	validator "gopkg.in/go-playground/validator.v8"
	"fmt"
)

// IsValidSubmittedAt validates the "SubmittedAt" value of FeedbackResponseSerializer
func IsValidSubmittedAt(
	v *validator.Validate,
	topStruct reflect.Value,
	currentStruct reflect.Value,
	field reflect.Value,
	fieldType reflect.Type,
	fieldKind reflect.Kind,
	param string,
) bool {
	feedbackStatus := currentStruct.Interface().(*feedbackSerializers.FeedbackResponseSerializer).Status
	if feedbackStatus == feedbackModels.SubmittedFeedback {
		// Check if the submitted at is in correct format
		_, err := time.Parse(time.RFC3339, field.String())
		if err != nil {
			return false
		}
	}
	return true
}

// IsAllQuestionPresent validates the "Data" value of FeedbackResponseSerializer
func IsAllQuestionPresent(db *gorm.DB) validator.Func {
	return func(
		v *validator.Validate,
		topStruct reflect.Value,
		currentStruct reflect.Value,
		field reflect.Value,
		fieldType reflect.Type,
		fieldKind reflect.Kind,
		param string,
	) bool {
		var expectedResponseIDs, actualResponseIDs []interface{}
		feedbackResponseData := currentStruct.Interface().(*feedbackSerializers.FeedbackResponseSerializer)
		if err := db.Model(feedbackModels.QuestionResponse{}).
			Where("feedback_id = ?", feedbackResponseData.FeedbackID).
			Pluck("id", &expectedResponseIDs).Error; err != nil {
			return false
		}
		for _, categoryData := range feedbackResponseData.Data {
			for _, skillData := range categoryData {
				for questionResponseID, questionResponseData := range skillData {
					if feedbackResponseData.Status == feedbackModels.SubmittedFeedback && questionResponseData.Response == "" {
						return false
					}
					actualResponseIDs = append(actualResponseIDs, questionResponseID)
				}
			}
		}
		expectedResponseIDSet := mapset.NewSetFromSlice(expectedResponseIDs)
		actualResponseIDSet := mapset.NewSetFromSlice(actualResponseIDs)
		return expectedResponseIDSet.Equal(actualResponseIDSet)
	}
}

func validators() {
	if err := binding.Validator.RegisterValidation("is_valid_submitted_at",
		IsValidSubmittedAt); err != nil {
		fmt.Println(err.Error())
	}
	if err := binding.Validator.RegisterValidation("all_questions_present",
		IsAllQuestionPresent(ctrl.FeedbackService.DB)); err != nil {
		fmt.Println(err.Error())
	}
}