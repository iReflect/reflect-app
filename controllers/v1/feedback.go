package v1

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	feedbackSerializers "github.com/iReflect/reflect-app/apps/feedback/serializers"
	feedbaclServices "github.com/iReflect/reflect-app/apps/feedback/services"
)

//FeedbackController ...
type FeedbackController struct {
	FeedbackService feedbaclServices.FeedbackService
}

// Routes for Feedback
func (ctrl FeedbackController) Routes(r *gin.RouterGroup) {
	r.GET("/", ctrl.List)
	r.GET("/:id/", ctrl.Get)
	r.PUT("/:id/", ctrl.Put)
}

// ToDo: handle errors like in retrospectives/sprints controllers

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
	userID, _ := c.Get("userID")
	feedBackResponseData := feedbackSerializers.FeedbackResponseSerializer{FeedbackID: id}
	if err := c.BindJSON(&feedBackResponseData); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid request data"})
		return
	}
	code, err := ctrl.FeedbackService.Put(id, userID.(uint), feedBackResponseData)
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
	perPage, _ := c.GetQuery("perPage")
	parsedPerPage, err := strconv.Atoi(perPage)
	if err != nil {
		parsedPerPage = -1
	}
	response, err := ctrl.FeedbackService.List(userID.(uint), statuses, parsedPerPage)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "Feedbacks not found", "error": err})
		return
	}
	c.JSON(http.StatusOK, response)
}
