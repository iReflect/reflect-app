package server

import (
	"net/http"

	"github.com/jinzhu/gorm"
	"github.com/qor/admin"
	"github.com/qor/qor"

	feedbackModels "github.com/iReflect/reflect-app/apps/feedback/models"
	userModels "github.com/iReflect/reflect-app/apps/user/models"
)

type Admin struct {
	DB *gorm.DB
}

func (a *Admin) Router() *http.ServeMux {

	adminRouter := http.NewServeMux()

	Admin := admin.New(&qor.Config{
		DB: a.DB,
	})

	category := feedbackModels.Category{}
	item := feedbackModels.Item{}
	itemType := feedbackModels.ItemType{}

	Admin.AddResource(&category, &admin.Config{Menu: []string{"Data Management"}})
	Admin.AddResource(&item, &admin.Config{Menu: []string{"Data Management"}})
	Admin.AddResource(&itemType, &admin.Config{Menu: []string{"Data Management"}})

	user := userModels.User{}
	group := userModels.Group{}
	role := userModels.Role{}

	Admin.AddResource(&user, &admin.Config{Menu: []string{"User Management"}})
	Admin.AddResource(&group, &admin.Config{Menu: []string{"User Management"}})
	Admin.AddResource(&role, &admin.Config{Menu: []string{"User Management"}})

	Admin.MountTo("/admin/", adminRouter)

	return adminRouter
}
