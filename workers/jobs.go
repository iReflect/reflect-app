package workers

import (
	"github.com/garyburd/redigo/redis"
	"github.com/gocraft/work"
	"github.com/iReflect/reflect-app/config"
	"log"
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

// Enqueuer ...
var Enqueuer = work.NewEnqueuer(redisNamespace, redisPool)

// Config ...
var Config *config.Config

// Pool ...
var Pool *work.WorkerPool

type job struct {
	name     string
	function func(*work.Job) error
}

var jobs []job

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
	Pool.Middleware(Log)

	assignJobs()

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
func Log(job *work.Job, next work.NextMiddlewareFunc) error {
	log.Println("Starting job: ", job.Name)
	return next()
}

func assignJobs() {
	// Map the name of jobs to handler functions
	for _, job := range jobs {
		Pool.Job(job.name, job.function)
	}
}

// RegisterJob ...
func RegisterJob(name string, function func(*work.Job) error) {
	jobs = append(jobs, job{name: name, function: function})
}
