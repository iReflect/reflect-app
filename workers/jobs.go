package workers

import (
	"errors"
	"github.com/garyburd/redigo/redis"
	"github.com/gocraft/work"
	"github.com/iReflect/reflect-app/config"
	"github.com/iReflect/reflect-app/db"
	"log"

	retroServices "github.com/iReflect/reflect-app/apps/retrospective/services"
)

var redisPool = &redis.Pool{
	MaxActive: 5,
	MaxIdle:   5,
	Wait:      true,
	Dial: func() (redis.Conn, error) {
		return redis.Dial("tcp", config.GetConfig().Redis.Address)
	},
}

// Workers ...
type Workers struct{}

var redisNamespace = "ireflect_worker"

var Enqueuer = work.NewEnqueuer(redisNamespace, redisPool)
var Config *config.Config
var Pool *work.WorkerPool

// Initialize ...
func (w *Workers) Initialize(config *config.Config) {
	Config = config

	// Make a new pool. Arguments:
	// Context{} is a struct that will be the context for the request.
	// 10 is the max concurrency
	// "my_app_namespace" is the Redis namespace
	// redisPool is a Redis pool
	Pool = work.NewWorkerPool(*w, 10, redisNamespace, redisPool)

	// Add middleware that will be executed for each job
	Pool.Middleware((*Workers).Log)

	w.assignJobs()

	// Start processing jobs
	Pool.Start()

	log.Println("Started Workers...")
}

// Shutdown ...
func (w *Workers) Shutdown() {
	// Stop the pool
	Pool.Stop()

	log.Println("Stopped Workers...")
}

// Log ...
func (w *Workers) Log(job *work.Job, next work.NextMiddlewareFunc) error {
	log.Println("Starting job: ", job.Name)
	return next()
}

// ToDo: Make generic
func (w *Workers) assignJobs() {
	// Map the name of jobs to handler functions
	Pool.Job("sync_sprint_data", (*Workers).SyncSprintData)
	Pool.Job("sync_sprint_member_data", (*Workers).SyncSprintMemberData)
}

// SyncSprintData ...
func (w *Workers) SyncSprintData(job *work.Job) error {
	sprintService := retroServices.SprintService{DB: db.Initialize(Config)}

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

// SyncSprintMemberData ...
func (w *Workers) SyncSprintMemberData(job *work.Job) error {
	sprintService := retroServices.SprintService{DB: db.Initialize(Config)}

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
