package middleware

import (
	"github.com/gin-gonic/gin"

	"github.com/jinzhu/gorm"
)

func DBMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("DB", db)
		c.Next()
	}
}
