package server

import (
	"net/http"

	"github.com/jinzhu/gorm"
	"github.com/qor/admin"
	"github.com/qor/qor"

	userModels "github.com/iReflect/reflect-app/apps/user/models"
	feedbackModels "github.com/iReflect/reflect-app/apps/feedback/models"
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
	Admin.AddResource(&userModels.Team{}, &admin.Config{Menu: []string{"User Management"}})
	Admin.AddResource(&userModels.UserTeamAssociation{}, &admin.Config{Menu: []string{"User Management"}})

	Admin.AddResource(&feedbackModels.Category{}, &admin.Config{Menu: []string{"Feedback Form Management"}})
	Admin.AddResource(&feedbackModels.Skill{}, &admin.Config{Menu: []string{"Feedback Form Management"}})
	Admin.AddResource(&feedbackModels.Question{}, &admin.Config{Menu: []string{"Feedback Form Management"}})
	Admin.AddResource(&feedbackModels.FeedbackForm{}, &admin.Config{Menu: []string{"Feedback Form Management"}})
	Admin.AddResource(&feedbackModels.FeedbackFormContent{}, &admin.Config{Menu: []string{"Feedback Form Management"}})
	Admin.AddResource(&feedbackModels.TeamFeedbackForm{}, &admin.Config{Menu: []string{"Feedback Form Management"}})

	Admin.AddResource(&feedbackModels.Feedback{}, &admin.Config{Menu: []string{"Feedback Management"}})
	Admin.AddResource(&feedbackModels.QuestionResponse{}, &admin.Config{Menu: []string{"Feedback Management"}})
	Admin.AddResource(&feedbackModels.ResponseComment{}, &admin.Config{Menu: []string{"Feedback Management"}})
	
	Admin.AddResource(&feedbackModels.Schedule{}, &admin.Config{Menu: []string{"Schedule Management"}})
	Admin.MountTo("/admin/", adminRouter)

	return adminRouter
}
