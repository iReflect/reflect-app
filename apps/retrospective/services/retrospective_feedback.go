package services

import (
	"github.com/iReflect/reflect-app/apps/retrospective/models"
	retrospectiveSerializers "github.com/iReflect/reflect-app/apps/retrospective/serializers"
	"github.com/iReflect/reflect-app/db"
	"github.com/iReflect/reflect-app/libs/utils"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"net/http"
	"strconv"
)

// RetrospectiveFeedbackService ...
type RetrospectiveFeedbackService struct {
}

// Add ...
func (service RetrospectiveFeedbackService) Add(userID uint, sprintID string, retroID string,
	feedbackType models.RetrospectiveFeedbackType,
	feedbackData *retrospectiveSerializers.RetrospectiveFeedbackCreateSerializer) (
	*retrospectiveSerializers.RetrospectiveFeedback,
	int,
	error) {
	db := db.DB

	retroIDInt, err := strconv.Atoi(retroID)
	if err != nil {
		return nil, http.StatusBadRequest, errors.New("invalid retrospective id")
	}
	sprint := models.Sprint{}

	if err := db.Model(&models.Sprint{}).
		Where("id = ?", sprintID).
		Find(&sprint).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, http.StatusNotFound, errors.New("sprint not found")
		}
		utils.LogToSentry(err)
		return nil, http.StatusInternalServerError, errors.New("failed to get sprint")
	}

	retroFeedback := models.RetrospectiveFeedback{
		RetrospectiveID: uint(retroIDInt),
		SubType:         feedbackData.SubType,
		Type:            feedbackType,
		AddedAt:         sprint.StartDate,
		CreatedByID:     userID,
		AssigneeID:      nil,
		ExpectedAt:      nil,
		ResolvedAt:      nil,
	}

	if feedbackType != models.GoalType {
		retroFeedback.ResolvedAt = sprint.EndDate
	}

	err = db.Create(&retroFeedback).Error
	if err != nil {
		return nil, http.StatusInternalServerError, errors.New("failed to get sprint")
	}

	return service.getRetrospectiveFeedback(retroFeedback.ID)

}

// Update ...
func (service RetrospectiveFeedbackService) Update(userID uint, retroID string,
	feedbackID string,
	feedbackData *retrospectiveSerializers.RetrospectiveFeedbackUpdateSerializer) (
	*retrospectiveSerializers.RetrospectiveFeedback,
	int,
	error) {
	db := db.DB

	retroFeedback := models.RetrospectiveFeedback{}

	if err := db.Model(&models.RetrospectiveFeedback{}).
		Where("id = ?", feedbackID).
		First(&retroFeedback).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, http.StatusNotFound, errors.New("retrospective feedback not found")
		}
		utils.LogToSentry(err)
		return nil, http.StatusInternalServerError, errors.New("failed to get retrospective feedback")
	}

	if retroFeedback.Type == models.GoalType && retroFeedback.ResolvedAt != nil {
		return nil, http.StatusBadRequest, errors.New("can not updated resolved goal")
	}

	if feedbackData.Scope != nil {
		retroFeedback.Scope = models.RetrospectiveFeedbackScope(*feedbackData.Scope)
	}

	if feedbackData.Text != nil {
		retroFeedback.Text = *feedbackData.Text
	}

	if feedbackData.ExpectedAt != nil {
		if retroFeedback.Type != models.GoalType {
			return nil, http.StatusBadRequest, errors.New("expectedAt can be updated only for goal " +
				"type retrospective feedback")
		}
		retroFeedback.ExpectedAt = feedbackData.ExpectedAt
	}

	retroFeedback.AssigneeID = feedbackData.AssigneeID

	err := db.Save(&retroFeedback).Error
	if err != nil {
		utils.LogToSentry(err)
		return nil, http.StatusInternalServerError, errors.New("failed to update retrospective feedback")
	}

	return service.getRetrospectiveFeedback(retroFeedback.ID)
}

// Resolve ...
func (service RetrospectiveFeedbackService) Resolve(userID uint, sprintID string, retroID string,
	feedbackID string,
	markResolved bool) (
	*retrospectiveSerializers.RetrospectiveFeedback,
	int,
	error) {
	db := db.DB

	retroFeedback := models.RetrospectiveFeedback{}

	sprint := models.Sprint{}

	if err := db.Model(&models.Sprint{}).
		Where("id = ?", sprintID).
		Find(&sprint).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, http.StatusNotFound, errors.New("sprint not found")
		}
		utils.LogToSentry(err)
		return nil, http.StatusInternalServerError, errors.New("failed to get sprint")
	}

	if err := db.Model(&models.RetrospectiveFeedback{}).
		Where("id = ?", feedbackID).
		First(&retroFeedback).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, http.StatusNotFound, errors.New("goal not found")
		}
		utils.LogToSentry(err)
		return nil, http.StatusInternalServerError, errors.New("failed to get goal")
	}

	if retroFeedback.Type != models.GoalType {
		return nil, http.StatusBadRequest, errors.New("only goal typed retrospective feedback could" +
			" be resolved or unresolved")
	}

	if markResolved && retroFeedback.ResolvedAt == nil {
		retroFeedback.ResolvedAt = sprint.EndDate
	}

	if !markResolved && retroFeedback.ResolvedAt != nil {
		retroFeedback.ResolvedAt = nil
	}

	err := db.Save(&retroFeedback).Error
	if err != nil {
		utils.LogToSentry(err)
		return nil, http.StatusInternalServerError, errors.New("failed to resolve goal")
	}

	return service.getRetrospectiveFeedback(retroFeedback.ID)
}

// List ...
func (service RetrospectiveFeedbackService) List(userID uint, sprintID string, retroID string,
	feedbackType models.RetrospectiveFeedbackType) (
	feedbackList *retrospectiveSerializers.RetrospectiveFeedbackListSerializer,
	status int,
	err error) {
	db := db.DB
	feedbackList = new(retrospectiveSerializers.RetrospectiveFeedbackListSerializer)
	sprint := models.Sprint{}

	if err := db.Model(&models.Sprint{}).
		Where("id = ?", sprintID).
		Find(&sprint).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, http.StatusNotFound, errors.New("sprint not found")
		}
		utils.LogToSentry(err)
		return nil, http.StatusInternalServerError, errors.New("failed to get sprint")
	}

	if err := db.Model(&models.RetrospectiveFeedback{}).
		Where("retrospective_id = ? AND type = ?", retroID, feedbackType).
		Where("added_at >= ? AND added_at <= ?", *sprint.StartDate, *sprint.EndDate).
		Preload("Assignee").
		Preload("CreatedBy").
		Order("added_at DESC, created_at DESC").
		Find(&feedbackList.Feedbacks).Error; err != nil {
		utils.LogToSentry(err)
		return nil, http.StatusInternalServerError, errors.New("failed to get retrospective feedbacks")
	}

	return feedbackList, http.StatusOK, nil
}

// ListGoal ...
func (service RetrospectiveFeedbackService) ListGoal(userID uint, sprintID string,
	retroID string, goalType string) (
	feedbackList *retrospectiveSerializers.RetrospectiveFeedbackListSerializer,
	status int,
	err error) {
	db := db.DB

	sprint := models.Sprint{}
	feedbackList = new(retrospectiveSerializers.RetrospectiveFeedbackListSerializer)

	if err := db.Model(&models.Sprint{}).
		Where("id = ?", sprintID).
		Find(&sprint).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, http.StatusNotFound, errors.New("sprint not found")
		}
		utils.LogToSentry(err)
		return nil, http.StatusInternalServerError, errors.New("failed to get sprint")
	}

	query := db.Model(&models.RetrospectiveFeedback{}).
		Where("retrospective_id = ? AND type = ?", retroID, models.GoalType)

	switch goalType {
	case "added":
		query = query.Where("resolved_at IS NULL").
			Where("added_at >= ? AND added_at <= ?", sprint.StartDate, sprint.EndDate).
			Order("added_at DESC, created_at DESC")
	case "completed":
		query = query.
			Where("resolved_at >= ? AND resolved_at <= ?", sprint.StartDate, sprint.EndDate).
			Order("resolved_at DESC, added_at DESC, created_at DESC")
	case "pending":
		query = query.
			Where("resolved_at IS NULL").
			Where("added_at < ?", sprint.EndDate).
			Order("expected_at, added_at DESC, created_at DESC")
	default:
		return nil, http.StatusBadRequest, errors.New("invalid goal type")
	}

	if err := query.
		Preload("Assignee").
		Preload("CreatedBy").
		Find(&feedbackList.Feedbacks).Error; err != nil {
		return nil, http.StatusInternalServerError, errors.New("failed to get goals")
	}
	return feedbackList, http.StatusOK, nil
}

func (service RetrospectiveFeedbackService) getRetrospectiveFeedback(retroFeedbackID uint) (
	*retrospectiveSerializers.RetrospectiveFeedback,
	int,
	error) {
	db := db.DB
	feedback := retrospectiveSerializers.RetrospectiveFeedback{}
	err := db.Model(&models.RetrospectiveFeedback{}).
		Where("id = ?", retroFeedbackID).
		Preload("CreatedBy").
		Preload("Assignee").
		First(&feedback).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, http.StatusNotFound, errors.New("sprint not found")
		}
		utils.LogToSentry(err)
		return nil, http.StatusInternalServerError, errors.New("failed to get sprint")
	}

	return &feedback, http.StatusOK, nil

}
