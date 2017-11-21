package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/iReflect/reflect-app/config"
	"github.com/iReflect/reflect-app/db"
	adminServer "github.com/iReflect/reflect-app/servers/admin"
	apiServer "github.com/iReflect/reflect-app/servers/api"

	_ "github.com/iReflect/reflect-app/db/migrations" //Init for all migrations
)

func main() {

	config := config.GetConfig()

	//Run migrations - Need to see how this would be possible with new goose.
	gormDB := db.Initialize(config)
	db.Migrate(config, gormDB)

	app := &apiServer.App{}
	app.Initialize(config)
	app.SetRoutes()
	srv := app.Server(":3000")

	adminApp := &adminServer.App{}
	adminApp.Initialize(config)
	adminApp.SetRoutes()
	adminSrv := adminApp.Server(":9000")

	go func() {
		// service connections
		if err := srv.ListenAndServe(); err != nil {
			log.Printf("listen: %s\n", err)
		}
	}()

	go func() {
		// service connections
		if err := adminSrv.ListenAndServe(); err != nil {
			log.Printf("listen: %s\n", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Println("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("App Server Shutdown:", err)
	}
	if err := adminSrv.Shutdown(ctx); err != nil {
		log.Fatal("Admin Server Shutdown:", err)
	}
	log.Println("Server exiting")

}
