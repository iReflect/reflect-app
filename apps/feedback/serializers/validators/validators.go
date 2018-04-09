package validators

import (
	"fmt"
	"github.com/gin-gonic/gin/binding"
	"github.com/iReflect/reflect-app/db"
)

//FeedbackValidators is used for registering validators for the feedback app
type FeedbackValidators struct {
}

// Register registers all the validators for the feedback serializers
func (validator FeedbackValidators) Register() {
	if err := binding.Validator.RegisterValidation("is_valid_submitted_at",
		IsValidSubmittedAt); err != nil {
		fmt.Println(err.Error())
	}

	if err := binding.Validator.RegisterValidation("all_questions_present",
		IsAllQuestionPresent(db.DB)); err != nil {
		fmt.Println(err.Error())
	}

	if err := binding.Validator.RegisterValidation("is_valid_question_response",
		IsValidQuestionResponse); err != nil {
		fmt.Println(err.Error())
	}
}
