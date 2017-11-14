package main

import (
	"database/sql"
	"flag"
	"log"
	"os"

	"github.com/iReflect/reflect-app/config"
	"github.com/pressly/goose"

	// Init DB drivers.
	_ "github.com/lib/pq"

	_ "github.com/iReflect/reflect-app/db/migrations"
)

var (
	flags = flag.NewFlagSet("goose", flag.ExitOnError)
)

func main() {

	flags.Usage = usage
	flags.Parse(os.Args[1:])

	args := flags.Args()

	if len(args) < 1 {
		flags.Usage()
		return
	}

	if args[0] == "-h" || args[0] == "--help" {
		flags.Usage()
		return
	}

	config := config.GetConfig()

	if len(args) > 1 && args[0] == "create" {
		if err := goose.Run("create", nil, config.DB.MigrationsDir, args[1:]...); err != nil {
			log.Fatalf("goose run: %v", err)
		}
		return
	}

	command := args[0]

	db, err := sql.Open(config.DB.Driver, config.DB.DSN)
	if err != nil {
		log.Fatalf("-dsn=%q: %v\n", config.DB.DSN, err)
	}

	arguments := []string{}
	if len(args) > 1 {
		arguments = append(arguments, args[3:]...)
	}

	if err := goose.Run(command, db, config.DB.MigrationsDir, arguments...); err != nil {
		log.Fatalf("goose run: %v", err)
	}
}

func usage() {
	log.Print(usagePrefix)
	flags.PrintDefaults()
	log.Print(usageCommands)
}

var (
	usagePrefix = `Usage: goose [OPTIONS] COMMAND

Examples:
    goose status
    goose create init sql
    goose create add_some_column sql
    goose create fetch_user_data go
    goose up

Options:
`
	usageCommands = `
Commands:
    up                   Migrate the DB to the most recent version available
    up-to VERSION        Migrate the DB to a specific VERSION
    down                 Roll back the version by 1
    down-to VERSION      Roll back to a specific VERSION
    redo                 Re-run the latest migration
    status               Dump the migration status for the current DB
    version              Print the current version of the database
    create NAME [sql|go] Creates new migration file with next version
`
)
