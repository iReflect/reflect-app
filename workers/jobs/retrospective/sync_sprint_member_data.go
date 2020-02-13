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
	workers.RegisterJob("sync_sprint_member_data", SyncSprintMemberData)
}

// SyncSprintMemberData ...
func SyncSprintMemberData(job *work.Job) error {
	sprintService := retroServices.SprintService{DB: db.Initialize(workers.Config)}

	sprintID := job.ArgInt64("sprintID")
	if sprintID == 0 {
		log.Println("Job failed: ", job.Name, " with error: sprintID cannot be zero")
		return errors.New("sprintID cannot be zero")
	}

	sprintMemberID := job.ArgString("sprintMemberID")
	if sprintMemberID == "" {
		log.Println("Job failed: ", job.Name, " with error: sprintMemberID cannot be blank")
		return errors.New("sprintMemberID cannot be blank")
	}

	err := sprintService.SyncSprintMemberData(uint(sprintID), sprintMemberID)

	if err != nil {
		log.Println("Job failed: ", job.Name, " with error: ", err)
		return err
	}

	log.Println("Completed job: ", job.Name)
	return nil
}
