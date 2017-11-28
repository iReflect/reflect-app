package server

import (
	"net/http"

	"github.com/jinzhu/gorm"
	"github.com/qor/admin"
	"github.com/qor/qor"

	userModels "github.com/iReflect/reflect-app/apps/user/models"
	feeddbackModels "github.com/iReflect/reflect-app/apps/feedback/models"
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

	Admin.AddResource(&feeddbackModels.Category{}, &admin.Config{Menu: []string{"Feedback Form Management"}})
	Admin.AddResource(&feeddbackModels.Skill{}, &admin.Config{Menu: []string{"Feedback Form Management"}})
	Admin.AddResource(&feeddbackModels.Question{}, &admin.Config{Menu: []string{"Feedback Form Management"}})
	Admin.AddResource(&feeddbackModels.FeedbackForm{}, &admin.Config{Menu: []string{"Feedback Form Management"}})
	Admin.AddResource(&feeddbackModels.FeedbackFormContent{}, &admin.Config{Menu: []string{"Feedback Form Management"}})
	Admin.AddResource(&feeddbackModels.TeamFeedbackForm{}, &admin.Config{Menu: []string{"Feedback Form Management"}})

	Admin.AddResource(&feeddbackModels.Feedback{}, &admin.Config{Menu: []string{"Feedback Management"}})
	Admin.AddResource(&feeddbackModels.QuestionResponse{}, &admin.Config{Menu: []string{"Feedback Management"}})
	Admin.AddResource(&feeddbackModels.ResponseComment{}, &admin.Config{Menu: []string{"Feedback Management"}})
	
	Admin.AddResource(&feeddbackModels.Schedule{}, &admin.Config{Menu: []string{"Schedule Management"}})
	Admin.MountTo("/admin/", adminRouter)

	return adminRouter
}
