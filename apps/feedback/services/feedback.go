package services

import (
	"errors"
	"net/http"
	"time"

	"github.com/jinzhu/gorm"

	feedbackModels "github.com/iReflect/reflect-app/apps/feedback/models"
	feedbackSerializers "github.com/iReflect/reflect-app/apps/feedback/serializers"
	userModels "github.com/iReflect/reflect-app/apps/user/models"
	"github.com/iReflect/reflect-app/constants"
)

//FeedbackService ...
type FeedbackService struct {
	DB *gorm.DB
}

// Get feedback by id
func (service FeedbackService) Get(feedbackID string, userID uint) (*feedbackSerializers.FeedbackDetailSerializer, error) {
	db := service.DB
	feedback := new(feedbackSerializers.FeedbackDetailSerializer)

	if err := db.Model(&feedbackModels.Feedback{}).
		Where("id = ?", feedbackID).
		Where("by_user_profile_id in (?)",
			db.Model(&userModels.UserProfile{}).Where("user_id = ?", userID).Select("id").QueryExpr()).
		Select("id, title, duration_start,duration_end, submitted_at, expire_at, status, feedback_form_id").
		Scan(&feedback).Error; err != nil {
		return nil, errors.New(constants.FeedbackWithIDNotFound)
	}

	return service.getFeedbackDetail(feedback)
}

// TeamGet feedback by id
func (service FeedbackService) TeamGet(feedbackID string, userID uint) (*feedbackSerializers.FeedbackDetailSerializer, error) {
	db := service.DB
	feedback := new(feedbackSerializers.FeedbackDetailSerializer)
	feedbackIds := service.getTeamFeedbackIDs(userID)

	if err := db.Model(&feedbackModels.Feedback{}).
		Where("id = ?", feedbackID).
		Where("id in (?)", feedbackIds).
		Select("id, title, duration_start,duration_end, submitted_at, expire_at, status, feedback_form_id").
		Scan(&feedback).Error; err != nil {
		return nil, errors.New(constants.FeedbackWithIDNotFound)
	}

	return service.getFeedbackDetail(feedback)
}

// List users Feedback
func (service FeedbackService) List(userID uint, statuses []string, perPage int) (*feedbackSerializers.FeedbackListSerializer, error) {
	db := service.DB
	baseQuery := db.Model(&feedbackModels.Feedback{}).
		Where("by_user_profile_id in (?)",
			db.Model(&userModels.UserProfile{}).Where("user_id = ?", userID).Select("id").QueryExpr())

	response, err := service.getFeedbackList(baseQuery, statuses, perPage)
	if err != nil {
		return nil, errors.New("no feedbacks found for the user")
	}
	return response, nil
}

// TeamList users Feedback
func (service FeedbackService) TeamList(userID uint, statuses []string, perPage int) (*feedbackSerializers.FeedbackListSerializer, error) {
	db := service.DB
	feedbackIds := service.getTeamFeedbackIDs(userID)
	baseQuery := db.Model(&feedbackModels.Feedback{}).
		Where("id in (?)", feedbackIds)

	response, err := service.getFeedbackList(baseQuery, statuses, perPage)
	if err != nil {
		return nil, errors.New("no feedbacks found for the team, which user is a part of")
	}
	return response, nil
}

func (service FeedbackService) getFeedbackList(baseQuery *gorm.DB, statuses []string, perPage int) (*feedbackSerializers.FeedbackListSerializer, error) {
	listQuery := baseQuery
	if len(statuses) > 0 {
		listQuery = listQuery.Where("status in (?)", statuses)
	}

	feedbacks := new(feedbackSerializers.FeedbackListSerializer)
	if err := listQuery.
		Preload("Team").
		Preload("ByUserProfile").
		Preload("ByUserProfile.User").
		Preload("ByUserProfile.Role").
		Preload("ForUserProfile").
		Preload("ForUserProfile.User").
		Preload("ForUserProfile.Role").
		Preload("FeedbackForm").
		Limit(perPage).
		Find(&feedbacks.Feedbacks).Error; err != nil {
		return nil, errors.New(constants.FeedbackNotFound)
	}
	baseQuery.Where("status = ?", feedbackModels.NewFeedback).Count(&feedbacks.NewFeedbackCount)
	baseQuery.Where("status = ?", feedbackModels.InProgressFeedback).Count(&feedbacks.DraftFeedbackCount)
	baseQuery.Where("status = ?", feedbackModels.SubmittedFeedback).Count(&feedbacks.SubmittedFeedbackCount)
	return feedbacks, nil
}

func (service FeedbackService) getFeedbackDetail(feedback *feedbackSerializers.FeedbackDetailSerializer) (*feedbackSerializers.FeedbackDetailSerializer, error) {
	db := service.DB
	feedBackFormContents := []feedbackModels.FeedbackFormContent{}

	if err := db.Model(&feedbackModels.FeedbackFormContent{}).
		Preload("Skill").
		Preload("Skill.Questions").
		Preload("Category").
		Group("id, category_id").
		Where("feedback_form_id in (?)", feedback.FeedbackFormID).
		Find(&feedBackFormContents).Error; err != nil {
		return nil, errors.New(constants.FeedbackFormNotReadyError)
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
			questionOptions := question.GetOptions()
			response := questionResponse.Response
			defaultValue, exists := questionOptions["defaultValue"].(string)
			if feedback.Status != feedbackModels.SubmittedFeedback && exists && response == "" {
				response = defaultValue
			}
			questionResponses = append(questionResponses,
				feedbackSerializers.QuestionResponseDetailSerializer{
					ID:         question.ID,
					Type:       question.Type,
					Text:       question.Text,
					Options:    questionOptions["values"],
					Weight:     question.Weight,
					ResponseID: questionResponse.ID,
					Response:   response,
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
func (service FeedbackService) Put(feedbackID string, userID uint,
	feedBackResponseData feedbackSerializers.FeedbackResponseSerializer) (int, error) {
	db := service.DB
	feedback := feedbackModels.Feedback{}
	// Find a feedback with the given ID which hasn't been submitted before
	if err := db.Model(&feedbackModels.Feedback{}).
		Where("id = ? AND status != ? AND expire_at >= ?", feedbackID, feedbackModels.SubmittedFeedback, time.Now()).
		Where("by_user_profile_id in (?)",
			db.Model(&userModels.UserProfile{}).Where("user_id = ?", userID).Select("id").QueryExpr()).
		First(&feedback).Error; err != nil {
		return http.StatusNotFound, errors.New(constants.FeedbackNotFound)
	}
	tx := db.Begin() // transaction begin
	for _, categoryData := range feedBackResponseData.Data {
		for _, skillData := range categoryData {
			for questionResponseID, questionResponseData := range skillData {
				questionResponse := feedbackModels.QuestionResponse{}
				if err := tx.Model(&feedbackModels.QuestionResponse{}).
					Where("id = ? AND feedback_id = ?", questionResponseID, feedbackID).
					Find(&questionResponse).Error; err != nil {
					// Roll back the transaction if any question response is not found
					tx.Rollback()
					return http.StatusBadRequest, errors.New(constants.QuestionNotFound)
				}
				questionResponse.Response = questionResponseData.Response
				questionResponse.Comment = questionResponseData.Comment
				if err := tx.Save(&questionResponse).Error; err != nil {
					// Roll back the transaction if any question response fails to update
					tx.Rollback()
					return http.StatusBadRequest, errors.New(constants.QuestionResponseUpdateError)
				}
			}
		}
	}
	if feedBackResponseData.Status == feedbackModels.SubmittedFeedback {
		submittedAt, _ := time.Parse(time.RFC3339, feedBackResponseData.SubmittedAt)
		if err := tx.Model(&feedback).Update(map[string]interface{}{
			"status":       feedBackResponseData.Status,
			"submitted_at": submittedAt,
		}).Error; err != nil {
			// Roll back the transaction if feedback status update fails to execute
			tx.Rollback()
			return http.StatusBadRequest, errors.New(constants.FeedbackUpdateError)
		}
	}
	tx.Commit() // transaction committed/end
	return http.StatusNoContent, nil
}

func (service FeedbackService) getTeamFeedbackIDs(userID uint) []uint {
	db := service.DB
	filterQuery := `
        SELECT id
        FROM feedbacks
        WHERE (team_id, for_user_profile_id) IN (SELECT
                                                    ut.team_id team_id,
                                                    up.id for_user_profile_id
                                                FROM user_teams ut
                                                    JOIN user_profiles up
                                                        ON ut.user_id = up.user_id
                                                WHERE ut.role = 0 AND ut.team_id IN (SELECT team_id
                                                                                    FROM user_teams
                                                                                    WHERE user_id = ? AND role = 1))
        UNION
        SELECT id
        FROM feedbacks
        WHERE (team_id, for_user_profile_id) IN (SELECT
                                                    ut.team_id team_id,
                                                    up.id for_user_profile_id
                                                FROM user_teams ut
                                                    JOIN user_profiles up
                                                        ON ut.user_id = up.user_id
                                                WHERE ut.team_id IN (SELECT team_id
                                                                     FROM user_teams
                                                                     WHERE user_id = ? AND role = 2))
        UNION
        SELECT id
        FROM feedbacks
        WHERE by_user_profile_id IN (SELECT id FROM user_profiles WHERE user_id = ?);
    `
	var feedbackIds []uint

	rows, _ := db.Raw(filterQuery, userID, userID, userID).Select("id").Rows()
	defer rows.Close()
	for rows.Next() {
		var feedbackID uint
		rows.Scan(&feedbackID)
		feedbackIds = append(feedbackIds, feedbackID)
	}
	return feedbackIds

}
