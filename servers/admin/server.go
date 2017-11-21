package admin

import (
	"net/http"

	feedbackModels "github.com/iReflect/reflect-app/apps/feedback/models"
	userModels "github.com/iReflect/reflect-app/apps/user/models"
	"github.com/iReflect/reflect-app/config"
	"github.com/iReflect/reflect-app/db"
	"github.com/jinzhu/gorm"
	"github.com/qor/admin"
	"github.com/qor/qor"
)

type App struct {
	DB     *gorm.DB
	Router *http.ServeMux
}

// Initialize initializes the app with predefined configuration
func (a *App) Initialize(config *config.Config) {

	a.DB = db.Initialize(config)
	a.Router = http.NewServeMux()

}

func (a *App) SetRoutes() {

	r := a.Router

	rootHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/admin/", http.StatusMovedPermanently)
	})
	r.Handle("/", rootHandler)

	// Initalize Admin app
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

	Admin.MountTo("/admin/", r)
}

func (a *App) Server(host string) *http.Server {
	r := a.Router

	srv := &http.Server{
		Addr:    host,
		Handler: r,
	}

	return srv
}
