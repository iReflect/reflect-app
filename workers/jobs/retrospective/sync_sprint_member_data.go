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

	sprintMemberID := job.ArgString("sprintMemberID")
	if sprintMemberID == "" {
		log.Println("Job failed: ", job.Name, " with error: sprintMemberID cannot be blank")
		return errors.New("sprintMemberID cannot be blank")
	}
	err := sprintService.SyncSprintMemberData(sprintMemberID)

	if err != nil {
		log.Println("Job failed: ", job.Name, " with error: ", err)
		return err
	}

	log.Println("Completed job: ", job.Name)
	return nil
}
