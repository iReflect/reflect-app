package services

import (
	"github.com/jinzhu/gorm"

	feedbackModels "github.com/iReflect/reflect-app/apps/feedback/models"
	feedbackSerializers "github.com/iReflect/reflect-app/apps/feedback/serializers"
	userModels "github.com/iReflect/reflect-app/apps/user/models"
)

//feedbackService ...
type FeedbackService struct {
	DB *gorm.DB
}

// Get feedback by id
func (service FeedbackService) Get(feedbackID string, userID uint) (feedback *feedbackSerializers.
	FeedbackDetailSerializer,
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

	categories := make(map[uint]feedbackSerializers.CategoryDetailSerializer)

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

// List users Feedback
func (service FeedbackService) List(userID uint, statuses []string) (
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
