package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/iReflect/reflect-app/config"
	"github.com/iReflect/reflect-app/db"
	"github.com/iReflect/reflect-app/servers"

	_ "github.com/iReflect/reflect-app/db/migrations" //Init for all migrations
	"github.com/iReflect/reflect-app/workers"
	_ "github.com/iReflect/reflect-app/workers/jobs/retrospective" // Init for jobs
)

func main() {

	configuration := config.GetConfig()

	//Run migrations - Need to see how this would be possible with new goose.
	db.Initialize(configuration)
	db.Migrate(configuration)

	app := &server.App{}
	app.Initialize(configuration)
	app.SetRoutes()
	app.SetAdminRoutes()
	srv := app.Server(":3000")

	go func() {
		// service connections
		if err := srv.ListenAndServe(); err != nil {
			log.Printf("listen: %s\n", err)
		}
	}()

	asyncWorkers := &workers.Workers{}
	asyncWorkers.Initialize(configuration)

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit

	asyncWorkers.Shutdown()

	log.Println("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("App Server Shutdown:", err)
	}
	log.Println("Server exiting")

}
