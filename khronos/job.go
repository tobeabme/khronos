package khronos

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	_ = iota
	// Success is status of a job whose last run was a success.
	Success
	// Running is status of a job whose last run has not finished.
	Running
	// Failed is status of a job whose last run was not successful on any nodes.
	Failed
	// PartialyFailed is status of a job whose last run was successful on only some nodes.
	PartialyFailed

	// ConcurrencyAllow allows a job to execute concurrency.
	ConcurrencyAllow = "allow"
	// ConcurrencyForbid forbids a job from executing concurrency.
	ConcurrencyForbid = "forbid"
)

type Job struct {
	running sync.Mutex

	//the job name must be unique in all of jobs
	Name string `json:"name"`

	// a breif description for job
	Breif string `json:"breif"`

	// e.g. “@every 1h30m10s” would indicate a schedule that activates every 1 hour, 30 minutes, 10 seconds.
	Schedule string `json:"schedule"`

	//local:shell,rpc remote:http
	JobType string `json:"job_type"`

	Shell bool

	// Command to run. Must be a shell command to execute.
	//shell e.g. "bash /path/to/my/script.sh" ;
	//rpc e.g. "handleFunc()"
	Command string `json:"command"`

	// for the remote job type
	HTTPProperties HTTPProperties `json:"http_properties"`

	//Job cannot run, as it is disabled
	Disabled bool `json:"disabled"`

	//the owner of this job
	Owner string `json:"owner"`

	// Email of the owner of this job
	// e.g. "admin@example.com"
	OwnerEmail string `json:"owner_email"`

	//allow (default): Allow concurrent job executions.
	//forbid: If the job is already running don’t send the execution, it will skip the executions until the next schedule.
	Concurrency string `json:"concurrency"`

	// Says if a job has been executed right numbers of time
	// and should not been executed again in the future
	IsDone bool `json:"is_done"`

	// Meta data about successful and failed runs.
	Metadata JobMetaData `json:"metadata"`

	//anything for job as a completement
	Payload map[string]string `json:"payload"`

	//Tags of this job.
	Tags map[string]string `json:"tags"`

	//job of Application.
	//the target servers of Application to run this job.
	Application string `json:"Application"`

	Agent *Agent `json:"-"`
}

type JobMetaData struct {
	SuccessCount uint `json:"success_count"`

	LastSuccess time.Time `json:"last_success"`

	ErrorCount uint `json:"error_count"`

	LastError time.Time `json:"last_error"`
}

// HTTPProperties Custom properties for the remote job type
type HTTPProperties struct {
	//e.g. http://localhost/jobHandle
	URL string `json:"url"`

	//GET POST PUT DELETE ...
	Method string `json:"method"`

	// A body to attach to the http request
	Body string `json:"body"`

	// A list of headers to add to http request (e.g. [{"key": "charset", "value": "UTF-8"}])
	Headers http.Header `json:"headers"`

	// A timeout property for the http request in seconds
	Timeout int `json:"timeout"`
}

// Run the job
func (j *Job) Run() {
	j.running.Lock()
	defer j.running.Unlock()
	// Maybe we are testing or it's disabled
	if j.Disabled == false {
		// Check if it's runnable
		if j.isRunnable() {
			log.WithFields(log.Fields{
				"job":         j.Name,
				"schedule":    j.Schedule,
				"jobType":     j.JobType,
				"application": j.Application,
			}).Debug("cron > job.Run: run a job")

			ex := NewExecution(j)
			ex.StartedAt = time.Now()
			j.Agent.Do(ex)
		}
	}
}

func (j *Job) isRunnable() bool {
	status := j.Status()

	if status == Running {
		if j.Concurrency == ConcurrencyAllow {
			return true
		} else if j.Concurrency == ConcurrencyForbid {
			log.WithFields(log.Fields{
				"job":         j.Name,
				"concurrency": j.Concurrency,
				"job_status":  status,
			}).Debug("scheduler > job.isRunnable: Skipping execution")
			return false
		}
	}

	return true
}

// Status returns the status of a job whether it's running, succeded or failed
func (j *Job) Status() int {
	execs, _ := j.Agent.store.GetLastExecutionGroup(j.Name)
	success := 0
	failed := 0
	for _, ex := range execs {
		if ex.FinishedAt.IsZero() {
			return Running
		}
	}

	var status int
	for _, ex := range execs {
		if ex.Success {
			success = success + 1
		} else {
			failed = failed + 1
		}
	}

	if failed == 0 {
		status = Success
	} else if failed > 0 && success == 0 {
		status = Failed
	} else if failed > 0 && success > 0 {
		status = PartialyFailed
	}

	return status
}

// Friendly format a job
func (j *Job) String() string {
	return fmt.Sprintf("\"name: %s, scheduled: %s, job_type: %s, disabled: %t, tags:%v\"", j.Name, j.Schedule, j.JobType, j.Disabled, j.Tags)
}
