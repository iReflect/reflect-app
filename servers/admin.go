package server

import (
	feedbackModels "github.com/iReflect/reflect-app/apps/feedback/models"
	retrospectiveModels "github.com/iReflect/reflect-app/apps/retrospective/models"
	userModels "github.com/iReflect/reflect-app/apps/user/models"
	"github.com/jinzhu/gorm"
	"github.com/qor/admin"
	"net/http"
)

type Admin struct {
	DB *gorm.DB
}

func (a *Admin) Router() *http.ServeMux {

	adminRouter := http.NewServeMux()

	Admin := admin.New(&admin.AdminConfig{
		SiteName: "Reflect Web",
		DB:       a.DB,
	})

	// User Management
	userModels.RegisterUserToAdmin(Admin, admin.Config{Menu: []string{"User Management"}})
	Admin.AddResource(&userModels.Role{}, &admin.Config{Menu: []string{"User Management"}})
	Admin.AddResource(&userModels.UserProfile{}, &admin.Config{Menu: []string{"User Management"}})
	userModels.RegisterTeamToAdmin(Admin, admin.Config{Menu: []string{"User Management"}})
	userModels.RegisterUserTeamToAdmin(Admin, admin.Config{Menu: []string{"User Management"}})

	// Retrospective Management
	retrospectiveModels.RegisterRetrospectiveToAdmin(Admin, admin.Config{Menu: []string{"Retrospective Management"}})
	retrospectiveModels.RegisterTaskToAdmin(Admin, admin.Config{Menu: []string{"Retrospective Management"}})
	retrospectiveModels.RegisterSprintToAdmin(Admin, admin.Config{Menu: []string{"Retrospective Management"}})
	Admin.AddResource(&retrospectiveModels.SprintSyncStatus{}, &admin.Config{Menu: []string{"Retrospective Management"}})
	retrospectiveModels.RegisterSprintMemberToAdmin(Admin, admin.Config{Menu: []string{"Retrospective Management"}})
	retrospectiveModels.RegisterSprintMemberTaskToAdmin(Admin, admin.Config{Menu: []string{"Retrospective Management"}})
	Admin.AddResource(&retrospectiveModels.RetrospectiveFeedback{}, &admin.Config{Menu: []string{"Retrospective Management"}})

	// Retrospective Audit Trails
	Admin.AddResource(&retrospectiveModels.Trail{}, &admin.Config{Menu: []string{"Retrospective Audit Trail Management"}})

	// Feedback Form Management
	Admin.AddResource(&feedbackModels.Category{}, &admin.Config{Menu: []string{"Feedback Form Management"}})
	feedbackModels.RegisterSkillToAdmin(Admin, admin.Config{Menu: []string{"Feedback Form Management"}})
	feedbackModels.RegisterQuestionToAdmin(Admin, admin.Config{Menu: []string{"Feedback Form Management"}})
	feedbackModels.RegisterFeedbackFormToAdmin(Admin, admin.Config{Menu: []string{"Feedback Form Management"}})
	Admin.AddResource(&feedbackModels.FeedbackFormContent{}, &admin.Config{Menu: []string{"Feedback Form Management"}})
	Admin.AddResource(&feedbackModels.TeamFeedbackForm{}, &admin.Config{Menu: []string{"Feedback Form Management"}})

	// Feedbacks Management
	feedbackModels.RegisterFeedbackToAdmin(Admin, admin.Config{Menu: []string{"Feedback Management"}})
	Admin.AddResource(&feedbackModels.QuestionResponse{}, &admin.Config{Menu: []string{"Feedback Management"}})

	// Schedule Management
	Admin.AddResource(&feedbackModels.Schedule{}, &admin.Config{Menu: []string{"Schedule Management"}})

	Admin.MountTo("/admin/", adminRouter)

	return adminRouter
}
