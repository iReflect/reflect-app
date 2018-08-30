package retrospective

import (
	"errors"
	"log"

	"github.com/gocraft/work"

	retroServices "github.com/iReflect/reflect-app/apps/retrospective/services"
	"github.com/iReflect/reflect-app/db"
	"github.com/iReflect/reflect-app/workers"
)

// AssignPointsToSprintTask ...
func AssignPointsToSprintTask(job *work.Job) error {
	sprintService := retroServices.SprintService{DB: db.Initialize(workers.Config)}

	sprintID := job.ArgString("sprintID")
	if sprintID == "" {
		log.Println("Job failed: ", job.Name, " with error: sprintID cannot be blank")
		return errors.New("sprintID cannot be blank")
	}

	sprintTaskID := job.ArgString("sprintTaskID")
	if sprintTaskID == "" {
		log.Println("Job failed: ", job.Name, " with error: sprintTaskID cannot be blank")
		return errors.New("sprintTaskID cannot be blank")
	}

	sprintService.AssignPoints(sprintID, &sprintTaskID)

	log.Println("Completed job: ", job.Name)
	return nil
}
