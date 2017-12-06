package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	feedbackModels "github.com/iReflect/reflect-app/apps/feedback/models"
	userModels "github.com/iReflect/reflect-app/apps/user/models"
	database "github.com/iReflect/reflect-app/db"
)

//FeedbackController ...
type FeedbackController struct{
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
	feedbacks := []feedbackModels.Feedback{}

	if err := db.Preload("Team").
		Preload("ByUserProfile").
		Preload("ByUserProfile.User").
		Preload("ByUserProfile.Role").
		Preload("ForUserProfile").
		Preload("ForUserProfile.User").
		Preload("ForUserProfile.Role").
		Preload("FeedbackForm").
		Where("by_user_profile_id in (?)",
		db.Model(&userModels.UserProfile{}).Where("user_id = ?", 1).Select("id").QueryExpr()).
		Find(&feedbacks).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "Feedbacks not found", "error": err})
		return
	}

	c.JSON(http.StatusOK, feedbacks)
}
