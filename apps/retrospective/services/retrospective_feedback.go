package services

import (
	"github.com/iReflect/reflect-app/apps/retrospective/models"
	retrospectiveSerializers "github.com/iReflect/reflect-app/apps/retrospective/serializers"
	"github.com/iReflect/reflect-app/apps/user/serializers"
	userServices "github.com/iReflect/reflect-app/apps/user/services"
	"github.com/jinzhu/gorm"
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
	*retrospectiveSerializers.RetrospectiveFeedbackSerializer,
	error) {
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
	err := db.Create(&retroFeedback).Error
	if err != nil {
		return nil, err
	}
	db.Model(&retroFeedback).Preload("CreatedBy")
	feedback := retrospectiveSerializers.RetrospectiveFeedbackSerializer{
		ID:              retroFeedback.ID,
		SubType:         retroFeedback.SubType,
		Type:            retroFeedback.Type,
		RetrospectiveID: retroFeedback.RetrospectiveID,
		Text:            retroFeedback.Text,
		Scope:           retroFeedback.Scope,
		Assignee:        nil,
		AddedAt:         retroFeedback.AddedAt,
		ResolvedAt:      retroFeedback.ResolvedAt,
		ExpectedAt:      retroFeedback.ExpectedAt,
		CreatedBy: serializers.User{
			ID:        retroFeedback.CreatedByID,
			Email:     retroFeedback.CreatedBy.Email,
			FirstName: retroFeedback.CreatedBy.FirstName,
			LastName:  retroFeedback.CreatedBy.LastName,
			Active:    retroFeedback.CreatedBy.Active,
		},
	}
	return &feedback, nil
}
