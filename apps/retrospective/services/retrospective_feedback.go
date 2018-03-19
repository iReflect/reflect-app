package services

import (
	"github.com/iReflect/reflect-app/apps/retrospective/models"
	retrospectiveSerializers "github.com/iReflect/reflect-app/apps/retrospective/serializers"
	userServices "github.com/iReflect/reflect-app/apps/user/services"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"strconv"
)

// RetrospectiveFeedbackService ...
type RetrospectiveFeedbackService struct {
	DB          *gorm.DB
	TeamService userServices.TeamService
}

// Add ...
func (service RetrospectiveFeedbackService) Add(userID uint, sprintID string, retroID string,
	feedbackType models.RetrospectiveFeedbackType,
	feedbackData *retrospectiveSerializers.RetrospectiveFeedbackCreateSerializer) (
	feedback *retrospectiveSerializers.RetrospectiveFeedback,
	err error) {
	db := service.DB

	retroIDInt, _ := strconv.Atoi(retroID)
	sprint := models.Sprint{}

	if err := db.Model(&models.Sprint{}).
		Where("id = ?", sprintID).
		Find(&sprint).Error; err != nil {
		return nil, err
	}

	feedback = new(retrospectiveSerializers.RetrospectiveFeedback)

	retroFeedback := models.RetrospectiveFeedback{
		RetrospectiveID: uint(retroIDInt),
		SubType:         feedbackData.SubType,
		Type:            feedbackType,
		AddedAt:         sprint.StartDate,
		CreatedByID:     userID,
		AssigneeID:      nil,
	}

	if feedbackType != models.GoalType {
		retroFeedback.ExpectedAt = sprint.EndDate
		retroFeedback.ResolvedAt = sprint.EndDate
	}

	err = db.Create(&retroFeedback).Error
	if err != nil {
		return nil, err
	}

	err = db.Model(&retroFeedback).Preload("CreatedBy").Error
	if err != nil {
		return nil, err
	}

	err = db.Model(&retroFeedback).Preload("CreatedBy").Preload("Assignee").
		First(feedback).Error
	if err != nil {
		return nil, err
	}

	return feedback, nil
}

// Update ...
func (service RetrospectiveFeedbackService) Update(userID uint, sprintID string, retroID string,
	feedbackID string,
	feedbackData *retrospectiveSerializers.RetrospectiveFeedbackUpdateSerializer) (
	feedback *retrospectiveSerializers.RetrospectiveFeedback,
	err error) {
	db := service.DB

	retroFeedback := models.RetrospectiveFeedback{}

	if err := db.Model(&models.RetrospectiveFeedback{}).
		Where("id = ?", feedbackID).
		First(&retroFeedback).Error; err != nil {
		return nil, err
	}

	if feedbackData.Scope != nil {
		retroFeedback.Scope = models.RetrospectiveFeedbackScope(*feedbackData.Scope)
	}

	if feedbackData.Text != nil {
		retroFeedback.Text = *feedbackData.Text
	}

	if feedbackData.ExpectedAt != nil {
		if retroFeedback.Type != models.GoalType {
			return nil, errors.New("expectedAt can be updated only for goal " +
				"type retrospective feedback")
		}
		retroFeedback.ExpectedAt = feedbackData.ExpectedAt
	}

	if feedbackData.AssigneeID != nil {
		var userID uint
		if err := db.Raw("SELECT user_teams.user_id FROM user_teams JOIN retrospectives "+
			"ON retrospectives.team_id = user_teams.team_id WHERE retrospectives.id = ? "+
			"and user_teams.user_id = ?;", retroID,
			feedbackData.AssigneeID).Scan(&userID).Error; err != nil {
			if err.Error() == "record not found" {
				return nil, errors.New("cannot assign to requesxted user")
			}
			return nil, err
		}
		retroFeedback.AssigneeID = feedbackData.AssigneeID
	}

	err = db.Save(&retroFeedback).Error
	if err != nil {
		return nil, err
	}

	err = db.Model(&retroFeedback).Preload("CreatedBy").Preload("Assignee").
		Scan(feedback).Error
	if err != nil {
		return nil, err
	}

	return feedback, nil
}

// Resolve ...
func (service RetrospectiveFeedbackService) Resolve(userID uint, sprintID string, retroID string,
	feedbackID string,
	markResolved bool) (
	feedback *retrospectiveSerializers.RetrospectiveFeedback,
	err error) {
	db := service.DB

	retroFeedback := models.RetrospectiveFeedback{}

	sprint := models.Sprint{}

	if err := db.Model(&models.Sprint{}).
		Where("id = ?", sprintID).
		Find(&sprint).Error; err != nil {
		return nil, err
	}

	if err := db.Model(&models.RetrospectiveFeedback{}).
		Where("id = ?", feedbackID).
		First(&retroFeedback).Error; err != nil {
		return nil, err
	}

	if retroFeedback.Type != models.GoalType {
		return nil, errors.New("only goal typed retrospective feedback could" +
			" be resolved or unresolved")
	}

	if markResolved && retroFeedback.ResolvedAt == nil {
		retroFeedback.ResolvedAt = sprint.EndDate
	}

	if !markResolved && retroFeedback.ResolvedAt != nil {
		retroFeedback.ResolvedAt = nil
	}

	err = db.Save(&retroFeedback).Error
	if err != nil {
		return nil, err
	}

	err = db.Model(&retroFeedback).Preload("CreatedBy").Preload("Assignee").
		Scan(feedback).Error
	if err != nil {
		return nil, err
	}

	return feedback, nil
}

// List ...
func (service RetrospectiveFeedbackService) List(userID uint, sprintID string, retroID string,
	feedbackType models.RetrospectiveFeedbackType) (
	feedbackList *retrospectiveSerializers.RetrospectiveFeedbackListSerializer,
	err error) {
	db := service.DB

	feedbackList = new(retrospectiveSerializers.RetrospectiveFeedbackListSerializer)
	sprint := models.Sprint{}

	if err := db.Model(&models.Sprint{}).
		Where("id = ?", sprintID).
		Find(&sprint).Error; err != nil {
		return nil, err
	}

	if err := db.Model(&models.RetrospectiveFeedback{}).
		Where("retrospective_id = ? AND type = ?", retroID, feedbackType).
		Where("added_at >= ? AND added_at < ?", *sprint.StartDate, *sprint.EndDate).
		Preload("Assignee").
		Preload("CreatedBy").
		Scan(&feedbackList.Feedbacks).Error; err != nil {
		return nil, err
	}

	return feedbackList, nil
}

// ListGoal ...
func (service RetrospectiveFeedbackService) ListGoal(userID uint, sprintID string,
	retroID string, goalType string) (
	feedbackList *retrospectiveSerializers.RetrospectiveFeedbackListSerializer,
	err error) {
	db := service.DB

	sprint := models.Sprint{}
	feedbackList = new(retrospectiveSerializers.RetrospectiveFeedbackListSerializer)

	if err := db.Model(&models.Sprint{}).
		Where("id = ?", sprintID).
		Find(&sprint).Error; err != nil {
		return nil, err
	}

	switch goalType {
	case "added":
		if err := db.Model(&models.RetrospectiveFeedback{}).
			Where("retrospective_id = ? AND type = ?", retroID, models.GoalType).
			Where("resolved_at IS NULL").
			Where("added_at >= ? AND added_at < ?", sprint.StartDate, sprint.EndDate).
			Preload("Assignee").
			Preload("CreatedBy").
			Scan(&feedbackList.Feedbacks).Error; err != nil {
			return nil, err
		}
	case "completed":
		if err := db.Model(&models.RetrospectiveFeedback{}).
			Where("retrospective_id = ? AND type = ?", retroID, models.GoalType).
			Where("resolved_at >= ? AND resolved_at < ?", sprint.StartDate, sprint.EndDate).
			Preload("Assignee").
			Preload("CreatedBy").
			Scan(&feedbackList.Feedbacks).Error; err != nil {
			return nil, err
		}
	case "pending":
		if err := db.Model(&models.RetrospectiveFeedback{}).
			Where("retrospective_id = ? AND type = ?", retroID, models.GoalType).
			Where("resolved_at IS NULL").
			Where("added_at < ?", sprint.EndDate).
			Preload("Assignee").
			Preload("CreatedBy").
			Scan(&feedbackList.Feedbacks).Error; err != nil {
			return nil, err
		}
	default:
		return nil, errors.New("invalid goal type")
	}
	return feedbackList, nil
}
