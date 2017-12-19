package services

import (
	"errors"
	"net/http"
	"time"

	"github.com/jinzhu/gorm"

	feedbackModels "github.com/iReflect/reflect-app/apps/feedback/models"
	feedbackSerializers "github.com/iReflect/reflect-app/apps/feedback/serializers"
	userModels "github.com/iReflect/reflect-app/apps/user/models"
)

//FeedbackService ...
type FeedbackService struct {
	DB *gorm.DB
}

// Get feedback by id
func (service FeedbackService) Get(feedbackID string, userID string) (feedback *feedbackSerializers.FeedbackDetailSerializer,
	err error) {
	db := service.DB
	feedback = new(feedbackSerializers.FeedbackDetailSerializer)

	if err := db.Model(&feedbackModels.Feedback{}).Where("id = ?", feedbackID).
		Where("by_user_profile_id in (?)",
			db.Model(&userModels.UserProfile{}).Where("user_id = ?", userID).Select("id").QueryExpr()).
		Select("id, title, duration_start,duration_end, submitted_at, expire_at, status, feedback_form_id").
		Scan(&feedback).Error; err != nil {
		return nil, err
	}

	feedBackFormContents := []feedbackModels.FeedbackFormContent{}

	if err := db.Model(&feedbackModels.FeedbackFormContent{}).
		Preload("Skill").
		Preload("Skill.Questions").
		Preload("Category").
		Group("id, category_id").
		Where("feedback_form_id in (?)", feedback.FeedbackFormID).
		Find(&feedBackFormContents).Error; err != nil {
		return nil, err
	}

	var categories = make(map[uint]feedbackSerializers.CategoryDetailSerializer)

	for _, feedBackFormContent := range feedBackFormContents {
		questionResponses := []feedbackSerializers.QuestionResponseDetailSerializer{}
		for _, question := range feedBackFormContent.Skill.Questions {
			questionResponse := feedbackModels.QuestionResponse{}
			db.Model(questionResponse).
				Where(feedbackModels.QuestionResponse{
					FeedbackID:            feedback.ID,
					QuestionID:            question.ID,
					FeedbackFormContentID: feedBackFormContent.ID,
				}).
				FirstOrCreate(&questionResponse)

			questionResponses = append(questionResponses,
				feedbackSerializers.QuestionResponseDetailSerializer{
					ID:         question.ID,
					Type:       question.Type,
					Text:       question.Text,
					Options:    question.Options,
					Weight:     question.Weight,
					ResponseID: questionResponse.ID,
					Response:   questionResponse.Response,
					Comment:    questionResponse.Comment,
				})
		}

		skill := feedbackSerializers.SkillDetailSerializer{
			ID:           feedBackFormContent.SkillID,
			Title:        feedBackFormContent.Skill.Title,
			DisplayTitle: feedBackFormContent.Skill.DisplayTitle,
			Description:  feedBackFormContent.Skill.Description,
			Weight:       feedBackFormContent.Skill.Weight,
			Questions:    questionResponses,
		}

		categoryID := feedBackFormContent.CategoryID
		_, exists := categories[categoryID]
		if exists == false {
			skills := make(map[uint]feedbackSerializers.SkillDetailSerializer)
			skills[feedBackFormContent.SkillID] = skill

			categories[categoryID] = feedbackSerializers.CategoryDetailSerializer{
				ID:          feedBackFormContent.Category.ID,
				Title:       feedBackFormContent.Category.Title,
				Description: feedBackFormContent.Category.Description,
				Skills:      skills,
			}
		} else {
			categories[categoryID].Skills[feedBackFormContent.SkillID] = skill
		}
	}
	feedback.Categories = categories
	return feedback, nil
}

// Put feedback data
func (service FeedbackService) Put(feedbackID string, userID string,
	feedBackResponseData feedbackSerializers.FeedbackResponseSerializer) (code int, err error) {
	db := service.DB
	feedback := feedbackModels.Feedback{}
	// Find a feedback with the given ID which hasn't been submitted before
	if err := db.Model(&feedbackModels.Feedback{}).Where("id = ? AND status != ?", feedbackID, 2).
		Where("by_user_profile_id in (?)",
			db.Model(&userModels.UserProfile{}).Where("user_id = ?", userID).Select("id").QueryExpr()).
		First(&feedback).Error; err != nil {
		code = http.StatusNotFound
		return code, err
	}
	tx := db.Begin() // transaction begin
	for _, categoryData := range feedBackResponseData.Data {
		for _, skillData := range categoryData {
			for questionResponseID, questionResponseData := range skillData {
				if rowsAffected := tx.Model(&feedbackModels.QuestionResponse{}).
					Where("id = ? AND feedback_id = ?", questionResponseID, feedbackID).
					Update(map[string]interface{}{
						"response": questionResponseData.Response,
						"comment":  questionResponseData.Comment,
					}).RowsAffected; rowsAffected == 0 {
					// Roll back the transaction if any question fails to execute
					tx.Rollback()
					code = http.StatusBadRequest
					err := errors.New("Failed to update the question response")
					return code, err
				}
			}
		}
	}
	if feedBackResponseData.SaveAndSubmit && feedBackResponseData.Status == 2 {
		if err := tx.Model(&feedback).Update(map[string]interface{}{
			"status":       2,
			"submitted_at": time.Now(),
		}).Error; err != nil {
			// Roll back the transaction if feedback status update fails to execute
			tx.Rollback()
			code = http.StatusBadRequest
			return code, err
		}
	}
	tx.Commit() // transaction committed/end
	return http.StatusNoContent, nil
}

// List users Feedback
func (service FeedbackService) List(userID string, statuses []string) (
	feedbacks *feedbackSerializers.FeedbackListSerializer,
	err error) {
	db := service.DB
	baseQuery := db.Model(&feedbackModels.Feedback{}).
		Where("by_user_profile_id in (?)",
			db.Model(&userModels.UserProfile{}).Where("user_id = ?", userID).Select("id").QueryExpr())

	listQuery := baseQuery
	if len(statuses) > 0 {
		listQuery = listQuery.Where("status in (?)", statuses)
	}

	feedbacks = new(feedbackSerializers.FeedbackListSerializer)
	if err := listQuery.
		Preload("Team").
		Preload("ByUserProfile").
		Preload("ByUserProfile.User").
		Preload("ByUserProfile.Role").
		Preload("ForUserProfile").
		Preload("ForUserProfile.User").
		Preload("ForUserProfile.Role").
		Preload("FeedbackForm").
		Find(&feedbacks.Feedbacks).Error; err != nil {
		return nil, err
	}
	baseQuery.Where("status = 0").Count(&feedbacks.NewFeedbackCount)
	baseQuery.Where("status = 1").Count(&feedbacks.DraftFeedbackCount)
	baseQuery.Where("status = 2").Count(&feedbacks.SubmittedFeedbackCount)
	return feedbacks, nil
}
