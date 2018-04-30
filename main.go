package main

import (
	"github.com/iReflect/reflect-app/commands"
	_ "github.com/iReflect/reflect-app/db/migrations"              //Init for all migrations
	_ "github.com/iReflect/reflect-app/workers/jobs/retrospective" // Init for jobs
)

func main() {
	commands.Execute()
}
