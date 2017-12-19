package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
	feedbackServices "github.com/iReflect/reflect-app/apps/feedback/services"
)

//FeedbackController ...
type FeedbackController struct {
	FeedbackService feedbackServices.FeedbackService
}

// Routes for Feedback
func (ctrl FeedbackController) Routes(r *gin.RouterGroup) {
	r.GET("/", ctrl.List)
	r.GET("/:id/", ctrl.Get)
	r.PUT("/:id/", ctrl.Put)
}

// Get feedback
func (ctrl FeedbackController) Get(c *gin.Context) {
	id := c.Param("id")
	userID, _ := c.Get("userID")
	feedbackResponse, err := ctrl.FeedbackService.Get(id, userID.(uint))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "Feedback not found", "error": err})
		return
	}

	c.JSON(http.StatusOK, feedbackResponse)
}

// Put feedback
func (ctrl FeedbackController) Put(c *gin.Context) {
	id := c.Param("id")
	feedBackResponseData := feedbackSerializers.FeedbackResponseSerializer{}
	if err := c.BindJSON(&feedBackResponseData); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "Invalid request data", "error": err.Error()})
		return
	}
	code, err := ctrl.FeedbackService.Put(id, "1", feedBackResponseData)
	if err != nil {
		c.AbortWithStatusJSON(code, gin.H{"message": "Error while saving the form!!", "error": err.Error()})
		return
	}
	c.JSON(code, nil)
}

// List Feedbacks
func (ctrl FeedbackController) List(c *gin.Context) {
	statuses := c.QueryArray("status")
	userID, _ := c.Get("userID")
	response, err := ctrl.FeedbackService.List(userID.(uint), statuses)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "Feedbacks not found", "error": err})
		return
	}
	c.JSON(http.StatusOK, response)
}
