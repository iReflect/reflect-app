package controllers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/jinzhu/gorm"

	"github.com/gin-gonic/gin"

	"github.com/iReflect/reflect-app/apps/feedback/models"
	"github.com/iReflect/reflect-app/db"
)

//ArticleController ...
type ItemController struct{}

//Add Routes
func (ctrl ItemController) Routes(r *gin.RouterGroup) {
	r.GET("/", ctrl.All)
	r.GET("/:id", ctrl.One)
}

//Get One ...
func (ctrl ItemController) One(c *gin.Context) {
	db, _ := db.GetFromContext(c)

	categoryTitle := c.Param("title")
	category := models.Category{}
	if err := db.First(&category, models.Category{Title: categoryTitle}).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "Category not found", "error": err})
		return
	}

	itemID, _ := strconv.Atoi(c.Param("id"))
	item := models.Item{Model: gorm.Model{ID: uint(itemID)}}
	if err := db.Model(&category).Related(&item).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "Item not found", "error": err})
		return
	}

	c.JSON(http.StatusOK, item)
}

//Get One ...
func (ctrl ItemController) All(c *gin.Context) {
	db, _ := db.GetFromContext(c)

	categoryTitle := c.Param("title")
	category := models.Category{}
	if err := db.First(&category, models.Category{Title: categoryTitle}).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "Category not found", "error": err})
		return
	}

	items := []models.Item{}
	if err := db.Preload("Category").Preload("ItemType").Model(&category).Related(&items).Error; err != nil {
		log.Printf("Item query failed: %s\n", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Item query failed"})
		return
	}
	c.JSON(http.StatusOK, items)
}
