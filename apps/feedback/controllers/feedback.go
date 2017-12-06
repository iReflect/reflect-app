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

//Add Routes
func (ctrl FeedbackController) Routes(r *gin.RouterGroup) {
	r.GET("/", ctrl.List)
	//r.GET("/:id", ctrl.Get)
	//r.PUT("/:id", ctrl.Update)
}

//Get feedback 
//func (ctrl FeedbackController) Get(c *gin.Context) {
//	id := c.Param("id")
//	db, _ := database.GetFromContext(c)
//	feedback := feedbackModels.Feedback{}
//
//	if err := db.First(&feedback, id).Error; err != nil {
//		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "Feedback not found", "error": err})
//		return
//	}
//	c.JSON(http.StatusOK, feedback)
//}

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
