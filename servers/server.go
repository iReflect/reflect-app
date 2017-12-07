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

	feedbackControllers "github.com/iReflect/reflect-app/apps/feedback/controllers"
	userControllers "github.com/iReflect/reflect-app/apps/user/controllers"
	"github.com/iReflect/reflect-app/config"
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
	r.Use(sessions.Sessions("session", store))

	r.Use(cors.Default())

	// Middleware
	r.Use(dbMiddlewares.DBMiddleware(a.DB))
}

func (a *App) SetRoutes() {
	r := a.Router

	v1 := r.Group("/api/v1")
	new(feedbackControllers.FeedbackController).Routes(v1.Group("feedbacks"))
	new(userControllers.UserAuthController).Routes(r.Group("/"))
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
