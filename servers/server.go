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

	feedbackValidators "github.com/iReflect/reflect-app/apps/feedback/serializers/validators"
	feedbackServices "github.com/iReflect/reflect-app/apps/feedback/services"
	retrospectiveValidators "github.com/iReflect/reflect-app/apps/retrospective/serializers/validators"
	retrospectiveServices "github.com/iReflect/reflect-app/apps/retrospective/services"
	_ "github.com/iReflect/reflect-app/apps/tasktracker/providers" // Register all the task-tracker providers
	taskTrackerServices "github.com/iReflect/reflect-app/apps/tasktracker/services"
	"github.com/iReflect/reflect-app/apps/user/middleware/oauth"
	userServices "github.com/iReflect/reflect-app/apps/user/services"
	"github.com/iReflect/reflect-app/config"
	"github.com/iReflect/reflect-app/controllers"
	apiControllers "github.com/iReflect/reflect-app/controllers/v1"
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
	store.Options(sessions.Options{HttpOnly: false, MaxAge: 4 * 60 * 60, Path: "/"})
	r.Use(sessions.Sessions("session", store))

	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = []string{"http://localhost:4200", "http://localhost:3000"}
	corsConfig.AllowCredentials = true
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "HEAD", "DELETE", "OPTIONS"}
	r.Use(cors.New(corsConfig))

	// Middleware
	r.Use(dbMiddlewares.DBMiddleware(a.DB))
}

func (a *App) SetRoutes() {
	r := a.Router

	authenticationService := userServices.AuthenticationService{DB: a.DB}

	v1 := r.Group("/api/v1")

	v1.Use(oauth.CookieAuthenticationMiddleWare(authenticationService))

	feedbackValidator := feedbackValidators.FeedbackValidators{DB: a.DB}
	feedbackValidator.Register()

	retrospectiveValidator := retrospectiveValidators.RetrospectiveValidators{DB: a.DB}
	retrospectiveValidator.Register()

	feedbackService := feedbackServices.FeedbackService{DB: a.DB}
	feedbackController := apiControllers.FeedbackController{FeedbackService: feedbackService}
	feedbackController.Routes(v1.Group("feedbacks"))

	teamFeedbackController := apiControllers.TeamFeedbackController{FeedbackService: feedbackService}
	teamFeedbackController.Routes(v1.Group("team-feedbacks"))

	userController := apiControllers.UserController{}
	userController.Routes(v1.Group("users"))

	teamService := userServices.TeamService{DB: a.DB}
	teamControllerRoute := v1.Group("teams")
	teamController := apiControllers.TeamController{TeamService: teamService}
	teamController.Routes(teamControllerRoute)

	authController := controllers.UserAuthController{AuthService: authenticationService}
	authController.Routes(r.Group("/"))

	permissionService := retrospectiveServices.PermissionService{DB: a.DB}
	trailService := retrospectiveServices.TrailService{DB: a.DB}
	retrospectiveService := retrospectiveServices.RetrospectiveService{DB: a.DB}
	retrospectiveRoute := v1.Group("retrospectives")

	retrospectiveController := apiControllers.RetrospectiveController{RetrospectiveService: retrospectiveService, PermissionService: permissionService, TrailService: trailService}
	retrospectiveController.Routes(retrospectiveRoute)

	sprintRoute := retrospectiveRoute.Group(":retroID/sprints")
	sprintService := retrospectiveServices.SprintService{DB: a.DB}
	sprintController := apiControllers.SprintController{SprintService: sprintService, PermissionService: permissionService, TrailService: trailService}
	sprintController.Routes(sprintRoute)

	taskService := retrospectiveServices.TaskService{DB: a.DB}
	taskRoute := sprintRoute.Group(":sprintID/tasks")
	tasksController := apiControllers.TaskController{TaskService: taskService, PermissionService: permissionService, TrailService: trailService}
	tasksController.Routes(taskRoute)

	taskTrackerService := taskTrackerServices.TaskTrackerService{}
	taskTrackerController := apiControllers.TaskTrackerController{TaskTrackerService: taskTrackerService}
	taskTrackerController.Routes(v1.Group("task-tracker"))
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
