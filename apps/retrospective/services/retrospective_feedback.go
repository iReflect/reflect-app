package services

import (
	"fmt"
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
	feedback *retrospectiveSerializers.RetrospectiveFeedbackSerializer,
	err error) {
	db := service.DB

	retroIDInt, _ := strconv.Atoi(retroID)
	sprint := models.Sprint{}

	if err := db.Model(&models.Sprint{}).
		Where("id = ?", sprintID).
		Find(&sprint).Error; err != nil {
		return nil, err
	}

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
		Scan(feedback).Error
	if err != nil {
		return nil, err
	}

	return feedback, nil
}

// Update ...
func (service RetrospectiveFeedbackService) Update(userID uint, sprintID string, retroID string,
	feedbackID string,
	feedbackData *retrospectiveSerializers.RetrospectiveFeedbackUpdateSerializer) (
	feedback *retrospectiveSerializers.RetrospectiveFeedbackSerializer,
	err error) {
	db := service.DB

	sprint := models.Sprint{}

	if err := db.Model(&models.Sprint{}).
		Where("id = ?", sprintID).
		Find(&sprint).Error; err != nil {
		return nil, err
	}

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
				return nil, errors.New(fmt.Sprintf("cannot assigne to user"))
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
