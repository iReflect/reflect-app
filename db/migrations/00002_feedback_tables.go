package migrations

import (
	"database/sql"
	"github.com/pressly/goose"
	"github.com/jinzhu/gorm"

	"github.com/iReflect/reflect-app/apps/feedback/models"
)

// Define only the fields used in this migration and not full model.
type Category struct {
	gorm.Model
	Weight int
}

func init() {
	goose.AddMigration(Up00002, Down00002)
}

func Up00002(tx *sql.Tx) error {
	// This code is executed when the migration is applied.
	gormdb, err := gorm.Open("postgres", interface{}(tx).(gorm.SQLCommon))
	if err != nil {
		return err
	}

	gormdb.CreateTable(&models.Skill{}, &models.Category{}, &models.Question{})
	gormdb.CreateTable(&models.FeedbackForm{}, &models.FeedbackFormContent{}, &models.TeamFeedbackForm{})
	gormdb.CreateTable(&models.Feedback{}, &models.QuestionResponse{})
	gormdb.CreateTable(&models.Schedule{})

	gormdb.Model(&models.Question{}).AddForeignKey("skill_id", "skills(id)", "RESTRICT", "RESTRICT")

	gormdb.Model(&models.FeedbackFormContent{}).AddForeignKey("feedback_form_id", "feedback_forms(id)", "RESTRICT",
		"RESTRICT")
	gormdb.Model(&models.FeedbackFormContent{}).AddForeignKey("skill_id", "skills(id)", "RESTRICT", "RESTRICT")
	gormdb.Model(&models.FeedbackFormContent{}).AddForeignKey("category_id", "categories(id)", "RESTRICT", "RESTRICT")

	gormdb.Model(&models.TeamFeedbackForm{}).AddForeignKey("team_id", "teams(id)", "RESTRICT", "RESTRICT")
	gormdb.Model(&models.TeamFeedbackForm{}).AddForeignKey("for_role_id", "roles(id)", "RESTRICT", "RESTRICT")
	gormdb.Model(&models.TeamFeedbackForm{}).AddForeignKey("feedback_form_id", "feedback_forms(id)", "RESTRICT", "RESTRICT")
	gormdb.Model(&models.TeamFeedbackForm{}).AddUniqueIndex("unique_team_user_role", "team_id", "for_role_id")

	gormdb.Model(&models.Feedback{}).AddForeignKey("feedback_form_id", "feedback_forms(id)", "RESTRICT", "RESTRICT")
	gormdb.Model(&models.Feedback{}).AddForeignKey("for_user_id", "users(id)", "RESTRICT", "RESTRICT")
	gormdb.Model(&models.Feedback{}).AddForeignKey("by_user_id", "users(id)", "RESTRICT", "RESTRICT")
	gormdb.Model(&models.Feedback{}).AddForeignKey("team_id", "teams(id)", "RESTRICT", "RESTRICT")
	
	gormdb.Model(&models.QuestionResponse{}).AddForeignKey("feedback_form_content_id", "feedback_form_contents(id)",
		"RESTRICT", "RESTRICT")
	gormdb.Model(&models.QuestionResponse{}).AddForeignKey("feedback_id", "feedbacks(id)", "RESTRICT", "RESTRICT")
	gormdb.Model(&models.QuestionResponse{}).AddForeignKey("question_id", "questions(id)", "RESTRICT", "RESTRICT")
	
	gormdb.Model(&models.Schedule{}).AddForeignKey("team_id", "teams(id)", "RESTRICT", "RESTRICT")

	return nil
}

func Down00002(tx *sql.Tx) error {
	// This code is executed when the migration is rolled back.
	gormdb, err := gorm.Open("postgres", interface{}(tx).(gorm.SQLCommon))
	if err != nil {
		return err
	}

	gormdb.Model(&models.Schedule{}).RemoveForeignKey("team_id", "teams(id)")
	
	gormdb.Model(&models.QuestionResponse{}).RemoveForeignKey("question_id", "questions(id)")
	gormdb.Model(&models.QuestionResponse{}).RemoveForeignKey("feedback_id", "feedbacks(id)")
	gormdb.Model(&models.QuestionResponse{}).RemoveForeignKey("feedback_form_content_id", "feedback_form_contents(id)")

	gormdb.Model(&models.Feedback{}).RemoveForeignKey("team_id", "teams(id)")
	gormdb.Model(&models.Feedback{}).RemoveForeignKey("by_user_id", "users(id)")
	gormdb.Model(&models.Feedback{}).RemoveForeignKey("by_user_id", "users(id)")
	gormdb.Model(&models.Feedback{}).RemoveForeignKey("feedback_form_id", "feedback_forms(id)")

	gormdb.Model(&models.TeamFeedbackForm{}).RemoveIndex("unique_team_user_role")
	gormdb.Model(&models.TeamFeedbackForm{}).RemoveForeignKey("feedback_form_id", "feedback_forms(id)")
	gormdb.Model(&models.TeamFeedbackForm{}).RemoveForeignKey("for_role_id", "roles(id)")
	gormdb.Model(&models.TeamFeedbackForm{}).RemoveForeignKey("team_id", "teams(id)")

	gormdb.Model(&models.FeedbackFormContent{}).RemoveForeignKey("category_id", "categories(id)")
	gormdb.Model(&models.FeedbackFormContent{}).RemoveForeignKey("skill_id", "skills(id)")
	gormdb.Model(&models.FeedbackFormContent{}).RemoveForeignKey("feedback_form_id", "feedback_forms(id)")
	gormdb.Model(&models.FeedbackFormContent{}).RemoveForeignKey("feedback_form_id", "feedback_forms(id)")

	gormdb.Model(&models.Question{}).RemoveForeignKey("skill_id", "skills(id)")

	gormdb.DropTable(&models.Schedule{})
	gormdb.DropTable(&models.Feedback{}, &models.QuestionResponse{})
	gormdb.DropTable(&models.FeedbackForm{}, &models.FeedbackFormContent{}, &models.TeamFeedbackForm{})
	gormdb.DropTable(&models.Skill{}, &models.Category{}, &models.Question{})

	return nil
}
