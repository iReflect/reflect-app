package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	feedbackModels "github.com/iReflect/reflect-app/apps/feedback/models"
	userModels "github.com/iReflect/reflect-app/apps/user/models"
	database "github.com/iReflect/reflect-app/db"
)

//FeedbackController ...
type FeedbackController struct {
}

// Routes for Feedback
func (ctrl FeedbackController) Routes(r *gin.RouterGroup) {
	r.GET("/", ctrl.List)
	r.GET("/:id/", ctrl.Get)
	//r.PUT("/:id", ctrl.Update)
}

// Get feedback
func (ctrl FeedbackController) Get(c *gin.Context) {
	id := c.Param("id")
	db, _ := database.GetFromContext(c)
	feedbackResponse := feedbackModels.FeedbackDetailResponse{}
	if err := db.Model(&feedbackModels.Feedback{}).Where("id = ?", id).Select("id, title, duration_start, duration_end, submitted_at, expire_at, status, feedback_form_id").
		Scan(&feedbackResponse).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "Feedback not found", "error": err})
		return
	}

	feedBackFormContents := []feedbackModels.FeedbackFormContent{}

	if err := db.Model(&feedbackModels.FeedbackFormContent{}).
		Preload("Skill").
		Preload("Skill.Questions").
		Preload("Category").
		Group("id, category_id").
		Where("feedback_form_id in (?)", feedbackResponse.FeedbackFormID).
		Find(&feedBackFormContents).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "Feedback not found", "error": err})
	}

	var categories = make(map[uint]feedbackModels.CategorySkillQuestions)

	for _, feedBackFormContent := range feedBackFormContents {
		questionResponseList := []feedbackModels.QuestionResponseDetail{}
		for _, question := range feedBackFormContent.Skill.Questions {
			questionResponse := feedbackModels.QuestionResponse{}
			db.Model(questionResponse).
				Where(feedbackModels.QuestionResponse{
					FeedbackID:            feedbackResponse.ID,
					QuestionID:            question.ID,
					FeedbackFormContentID: feedBackFormContent.ID,
				}).
				FirstOrCreate(&questionResponse)
			questionResponseList = append(questionResponseList,
				feedbackModels.QuestionResponseDetail{
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

		skillQuestionResponse := feedbackModels.SkillQuestionList{
			ID:           feedBackFormContent.SkillID,
			Title:        feedBackFormContent.Skill.Title,
			DisplayTitle: feedBackFormContent.Skill.DisplayTitle,
			Description:  feedBackFormContent.Skill.Description,
			Weight:       feedBackFormContent.Skill.Weight,
			Questions:    questionResponseList,
		}

		categoryID := feedBackFormContent.CategoryID
		_, exists := categories[categoryID]
		if exists == false {
			skillQuestionMap := make(map[uint]feedbackModels.SkillQuestionList)
			skillQuestionMap[feedBackFormContent.SkillID] = skillQuestionResponse

			categories[categoryID] = feedbackModels.CategorySkillQuestions{
				ID:          feedBackFormContent.Category.ID,
				Title:       feedBackFormContent.Category.Title,
				Description: feedBackFormContent.Category.Description,
				Skills:      skillQuestionMap,
			}
		} else {
			categories[categoryID].Skills[feedBackFormContent.SkillID] = skillQuestionResponse
		}
	}
	c.JSON(http.StatusOK, feedbackModels.FeedbackDetailResponse{
		ID:            feedbackResponse.ID,
		Title:         feedbackResponse.Title,
		DurationStart: feedbackResponse.DurationStart,
		DurationEnd:   feedbackResponse.DurationEnd,
		SubmittedAt:   feedbackResponse.SubmittedAt,
		ExpireAt:      feedbackResponse.ExpireAt,
		Status:        feedbackResponse.Status,
		Categories:    categories,
	})
}

// List Feedbacks
func (ctrl FeedbackController) List(c *gin.Context) {
	db, _ := database.GetFromContext(c)
	status := c.QueryArray("status")
	response := feedbackModels.FeedbackListResponse{}
	baseQuery := db.Model(&feedbackModels.Feedback{}).
		Where("by_user_profile_id in (?)",
			db.Model(&userModels.UserProfile{}).Where("user_id = ?", 1).Select("id").QueryExpr())

	listQuery := baseQuery
	if len(status) > 0 {
		listQuery = listQuery.Where("status in (?)", status)

	}

	if err := listQuery.
		Preload("Team").
		Preload("ByUserProfile").
		Preload("ByUserProfile.User").
		Preload("ByUserProfile.Role").
		Preload("ForUserProfile").
		Preload("ForUserProfile.User").
		Preload("ForUserProfile.Role").
		Preload("FeedbackForm").
		Find(&response.Feedbacks).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "Feedbacks not found", "error": err})
		return
	}
	baseQuery.Where("status = 0").Count(&response.NewFeedbackCount)
	baseQuery.Where("status = 1").Count(&response.DraftFeedbackCount)
	baseQuery.Where("status = 2").Count(&response.SubmittedFeedbackCount)

	c.JSON(http.StatusOK, response)
}
