package validators

import (
	"fmt"
	"github.com/gin-gonic/gin/binding"
	"github.com/jinzhu/gorm"
	"gopkg.in/go-playground/validator.v8"
)

//FeedbackValidators is used for registering validators for the feedback app
type FeedbackValidators struct {
	DB *gorm.DB
}

// Register registers all the validators for the feedback serializers
func (feedbackValidator FeedbackValidators) Register() {

	validatorEngine := binding.Validator.Engine().(*validator.Validate)

	if err := validatorEngine.RegisterValidation("is_valid_submitted_at",
		IsValidSubmittedAt); err != nil {
		fmt.Println(err.Error())
	}

	if err := validatorEngine.RegisterValidation("all_questions_present",
		IsAllQuestionPresent(feedbackValidator.DB)); err != nil {
		fmt.Println(err.Error())
	}

	if err := validatorEngine.RegisterValidation("is_valid_question_response",
		IsValidQuestionResponse); err != nil {
		fmt.Println(err.Error())
	}
}
