package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
	feedbackModels "github.com/iReflect/reflect-app/apps/feedback/models"
	feedbackSerializers "github.com/iReflect/reflect-app/apps/feedback/serializers"
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
	feedbackResponse := feedbackSerializers.FeedbackDetailSerializer{}
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

	var categories = make(map[uint]feedbackSerializers.CategoryDetailSerializer)

	for _, feedBackFormContent := range feedBackFormContents {
		questionResponses := []feedbackSerializers.QuestionResponseDetailSerializer{}
		for _, question := range feedBackFormContent.Skill.Questions {
			questionResponse := feedbackModels.QuestionResponse{}
			db.Model(questionResponse).
				Where(feedbackModels.QuestionResponse{
				FeedbackID:            feedbackResponse.ID,
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
	c.JSON(http.StatusOK, feedbackSerializers.FeedbackDetailSerializer{
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
	response := feedbackSerializers.FeedbackListSerializer{}
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
	response.Token = state.(string)
	c.JSON(http.StatusOK, response)
}
