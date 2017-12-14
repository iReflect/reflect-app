package server

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/contrib/ginrus"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"

	feedbackServices "github.com/iReflect/reflect-app/apps/feedback/services"
	"github.com/iReflect/reflect-app/config"
	controllers "github.com/iReflect/reflect-app/controllers/v1"
	"github.com/iReflect/reflect-app/db"
	dbMiddlewares "github.com/iReflect/reflect-app/db/middlewares"
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

	store := sessions.NewCookieStore([]byte(config.Auth.Secret))
	store.Options(sessions.Options{HttpOnly: false})
	r.Use(sessions.Sessions("session", store))
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	corsConfig.AllowCredentials = true
	r.Use(cors.New(corsConfig))

	// Middleware
	r.Use(dbMiddlewares.DBMiddleware(a.DB))
}

func (a *App) SetRoutes() {
	r := a.Router
	feedBackService := feedbackServices.FeedbackService{DB: a.DB}
	feedbackController := controllers.FeedbackController{FeedbackService: feedBackService}
	v1 := r.Group("/api/v1")
	feedbackController.Routes(v1.Group("feedbacks"))
}

func (a *App) SetAdminRoutes() {
	r := a.Router
	admin := &Admin{DB: a.DB}
	adminRouter := admin.Router()

	r.Any("/admin/*w", gin.WrapH(adminRouter))
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
