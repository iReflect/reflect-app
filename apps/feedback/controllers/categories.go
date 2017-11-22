package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/iReflect/reflect-app/apps/feedback/models"
	"github.com/iReflect/reflect-app/db"
)

//ArticleController ...
type CategoryController struct{}

//Add Routes
func (ctrl CategoryController) Routes(r *gin.RouterGroup) {
	r.GET("/", ctrl.All)
	r.GET("/:title", ctrl.One)
}

//Get One ...
func (ctrl CategoryController) One(c *gin.Context) {
	title := c.Param("title")
	db, _ := db.GetFromContext(c)
	category := models.Category{}
	if err := db.First(&category, models.Category{Title: title}).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "Category not found", "error": err})
		return
	}
	c.JSON(http.StatusOK, category)
}

func (ctrl CategoryController) All(c *gin.Context) {
	db, _ := db.GetFromContext(c)
	categories := []models.Category{}
	db.Find(&categories)
	c.JSON(http.StatusOK, categories)
}
