package retrospective

import (
	"errors"
	"github.com/gocraft/work"
	"github.com/iReflect/reflect-app/db"
	"github.com/iReflect/reflect-app/workers"
	"log"

	retroServices "github.com/iReflect/reflect-app/apps/retrospective/services"
)

func init() {
	workers.RegisterJob("sync_sprint_data", SyncSprintData)
}

// SyncSprintData ...
func SyncSprintData(job *work.Job) error {
	sprintService := retroServices.SprintService{DB: db.Initialize(workers.Config)}

	sprintID := job.ArgString("sprintID")
	if sprintID == "" {
		log.Println("Job failed: ", job.Name, " with error: sprintID cannot be blank")
		return errors.New("sprintID cannot be blank")
	}

	err := sprintService.SyncSprintData(sprintID)

	if err != nil {
		log.Println("Job failed: ", job.Name, " with error: ", err)
		return err
	}

	log.Println("Completed job: ", job.Name)
	return nil
}
