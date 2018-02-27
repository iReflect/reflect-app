package server

import (
	feedbackModels "github.com/iReflect/reflect-app/apps/feedback/models"
	retrospectiveModels "github.com/iReflect/reflect-app/apps/retrospective/models"
	userModels "github.com/iReflect/reflect-app/apps/user/models"
	"github.com/jinzhu/gorm"
	"github.com/qor/admin"
	"github.com/qor/qor"
	"net/http"
)

type Admin struct {
	DB *gorm.DB
}

func (a *Admin) Router() *http.ServeMux {

	adminRouter := http.NewServeMux()

	Admin := admin.New(&qor.Config{
		DB: a.DB,
	})

	Admin.AddResource(&userModels.User{}, &admin.Config{Menu: []string{"User Management"}})
	Admin.AddResource(&userModels.Role{}, &admin.Config{Menu: []string{"User Management"}})
	Admin.AddResource(&userModels.UserProfile{}, &admin.Config{Menu: []string{"User Management"}})
	Admin.AddResource(&userModels.Team{}, &admin.Config{Menu: []string{"User Management"}})
	userModels.RegisterUserTeamToAdmin(Admin, admin.Config{Menu: []string{"User Management"}})

	Admin.AddResource(&feedbackModels.Category{}, &admin.Config{Menu: []string{"Feedback Form Management"}})
	feedbackModels.RegisterSkillToAdmin(Admin, admin.Config{Menu: []string{"Feedback Form Management"}})

	feedbackModels.RegisterQuestionToAdmin(Admin, admin.Config{Menu: []string{"Feedback Form Management"}})
	feedbackModels.RegisterFeedbackFormToAdmin(Admin, admin.Config{Menu: []string{"Feedback Form Management"}})
	Admin.AddResource(&feedbackModels.FeedbackFormContent{}, &admin.Config{Menu: []string{"Feedback Form Management"}})
	Admin.AddResource(&feedbackModels.TeamFeedbackForm{}, &admin.Config{Menu: []string{"Feedback Form Management"}})

	feedbackModels.RegisterFeedbackToAdmin(Admin, admin.Config{Menu: []string{"Feedback Management"}})
	Admin.AddResource(&feedbackModels.QuestionResponse{}, &admin.Config{Menu: []string{"Feedback Management"}})

	Admin.AddResource(&feedbackModels.Schedule{}, &admin.Config{Menu: []string{"Schedule Management"}})

	retrospectiveModels.RegisterRetrospectiveToAdmin(Admin, admin.Config{Menu: []string{"Retrospective Management"}})
	retrospectiveModels.RegisterTaskToAdmin(Admin, admin.Config{Menu: []string{"Retrospective Management"}})
	retrospectiveModels.RegisterSprintToAdmin(Admin, admin.Config{Menu: []string{"Retrospective Management"}})
	Admin.AddResource(&retrospectiveModels.SprintMember{}, &admin.Config{Menu: []string{"Retrospective Management"}})
	Admin.AddResource(&retrospectiveModels.SprintMemberTask{}, &admin.Config{Menu: []string{"Retrospective Management"}})
	Admin.AddResource(&retrospectiveModels.Trail{}, &admin.Config{Menu: []string{"Retrospective Audit Trail Management"}})
	//Todo: Fix SMT Creation

	Admin.MountTo("/admin/", adminRouter)

	return adminRouter
}
