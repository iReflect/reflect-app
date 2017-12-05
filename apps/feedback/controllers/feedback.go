package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	feedbackModels "github.com/iReflect/reflect-app/apps/feedback/models"
	userModels "github.com/iReflect/reflect-app/apps/user/models"
	database "github.com/iReflect/reflect-app/db"
)

//FeedbackController ...
type FeedbackController struct{}

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
	byUserProfiles := userModels.UserProfile{}
	feedbacks := []feedbackModels.Feedback{}
	if err := db.Find(&byUserProfiles, userModels.UserProfile{UserID: 1}).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "User not found", "error": err})
		return
	}

	if err := db.Preload("Team").
		Preload("ByUserProfile").
		Preload("ByUserProfile.User").
		Preload("ByUserProfile.Role").
		Preload("ForUserProfile").
		Preload("ForUserProfile.User").
		Preload("ForUserProfile.Role").
		Preload("FeedbackForm").
		Find(&feedbacks, feedbackModels.Feedback{ByUserProfileID: byUserProfiles.UserID}).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "Feedbacks not found", "error": err})
		return
	}

	c.JSON(http.StatusOK, feedbacks)
}
