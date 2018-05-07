package commands

import (
	"context"
	"fmt"
	"github.com/iReflect/reflect-app/config"
	"github.com/iReflect/reflect-app/db"
	"github.com/iReflect/reflect-app/servers"
	"github.com/iReflect/reflect-app/workers"
	"github.com/spf13/cobra"
	"log"
	"os"
	"os/signal"
	"time"
)

var rootCmd = &cobra.Command{
	Use:   "reflect-app",
	Short: "reflect-app is a retrospective and feedback tool",
	Long: `reflect-app is a retrospective and feedback tool
                Complete documentation is available at https://ireflect.github.io`,
	Run: func(cmd *cobra.Command, args []string) {
		runCommand(cmd, args)
	},
}

// Execute ...
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func runCommand(cmd *cobra.Command, args []string) {
	configuration := config.GetConfig()

	//Run migrations - Need to see how this would be possible with new goose.
	gormDB := db.Initialize(configuration)
	db.Migrate(configuration, gormDB)

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
