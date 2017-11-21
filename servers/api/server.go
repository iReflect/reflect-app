package api

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/contrib/ginrus"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	feedbackControllers "github.com/iReflect/reflect-app/apps/feedback/controllers"
	"github.com/iReflect/reflect-app/config"
	"github.com/iReflect/reflect-app/db"
	appMiddlewares "github.com/iReflect/reflect-app/db/middlewares"
	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"
)

type App struct {
	DB     *gorm.DB
	Router *gin.Engine
}

// Initialize initializes the app with predefined configuration
func (a *App) Initialize(config *config.Config) {

	a.DB = db.Initialize(config)

	// Creates a router without any middleware by default
	r := gin.New()

	a.Router = r

	log.Println("Started...")

	// Global middleware
	r.Use(ginrus.Ginrus(logrus.StandardLogger(), time.RFC3339, true))

	// Recovery middleware recovers from any panics and writes a 500 if there was one.
	r.Use(gin.Recovery())

	store := sessions.NewCookieStore([]byte("secret"))
	r.Use(sessions.Sessions("mysession", store))

	// Middleware
	r.Use(appMiddlewares.DBMiddleware(a.DB))
}

func (a *App) SetRoutes() {
	r := a.Router

	v1 := r.Group("/api/v1")
	{
		v1.Group("categories")
		new(feedbackControllers.CategoryController).Routes(v1.Group("categories"))
		new(feedbackControllers.ItemController).Routes(v1.Group("categories/:title/items"))
	}
}

// Run the app on it's router
func (a *App) Server(host string) *http.Server {
	r := a.Router

	srv := &http.Server{
		Addr:    host,
		Handler: r,
	}

	return srv
}
