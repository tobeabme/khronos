package khronos

import (
	"fmt"
	"sort"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

// Execution type holds all of the details of a specific Execution.
type Execution struct {
	mux sync.Mutex

	// Name of the job this executions refers to.
	JobName string `json:"job_name,omitempty"`

	Payload map[string]string `json:"payload,omitempty"`

	Tags map[string]string `json:"tags,omitempty"`

	// Start time of the execution.
	StartedAt time.Time `json:"started_at,omitempty"`

	// When the execution finished running.
	FinishedAt time.Time `json:"finished_at,omitempty"`

	// If this execution executed succesfully.
	Success bool `json:"success,omitempty"`

	// Partial output of the execution.
	Output []byte `json:"output,omitempty"`

	// Node name of the node that run this execution.
	NodeName string `json:"node_name,omitempty"`

	// Execution group to what this execution belongs to.
	// one job to many excutions for every scheduling.
	Group int64 `json:"group,omitempty"`

	Application string `json:"application,omitempty"`

	// Retry attempt of this execution.
	Attempt uint `json:"attempt,omitempty"`

	//allow (default): Allow concurrent job executions.
	//forbid: If the job is already running don’t send the execution, it will skip the executions until the next schedule.
	Concurrency string `json:"concurrency"`

	// *Job

}

// NewExecution creates a new execution.
func NewExecution(j *Job) *Execution {
	return &Execution{
		JobName:     j.Name,
		Payload:     j.Payload,
		Tags:        j.Tags,
		Application: j.Application,
		Group:       time.Now().UnixNano(),
		Concurrency: j.Concurrency,
		Attempt:     1,
		// Job:     j,
	}
}

// Key wil generate the execution Id for an execution.
func (e *Execution) Key() string {
	return fmt.Sprintf("%d-%s", e.StartedAt.UnixNano(), e.NodeName)
}

func (e *Execution) IncCounter(nodeName string, quota string) {
	c := NewCounter()
	c.Plus(nodeName, quota)

	// log.WithFields(log.Fields{
	// 	"nodeName": nodeName,
	// 	"quota":    quota,
	// 	"Gcounter": Gcounter,
	// }).Debug("Execution.IncCounter.")

}

func (e *Execution) DecCounter(nodeName string, quota string) {
	c := NewCounter()
	c.Minus(nodeName, quota)

	// log.WithFields(log.Fields{
	// 	"nodeName": nodeName,
	// 	"quota":    quota,
	// 	"Gcounter": Gcounter,
	// }).Debug("Execution.DecCounter.")

}

// processors = []*Processor{
// 	{
// 		Application: "spider",
// 		NodeName:    "server-002",
// 		IP:          "127.0.0.1",
// 		Port:        9002,
// 		Status:      true,
// 		Undone:      6,
// 	},
// 	{
// 		Application: "spider",
// 		NodeName:    "server-001",
// 		IP:          "127.0.0.1",
// 		Port:        9001,
// 		Status:      true,
// 		Undone:      3,
// 	},
// 	{
// 		Application: "spider",
// 		NodeName:    "server-003",
// 		IP:          "127.0.0.1",
// 		Port:        9003,
// 		Status:      true,
// 		Undone:      8,
// 	},
// }
func (e *Execution) CheckCounter(processors []*Processor) []*Processor {

	var undone int

	if len(processors) > 0 {
		for k, p := range processors {

			if e.Tags["type"] == "" {
				// point to all of undone jobs in worker servers
				undone = Gcounter.Get(p.NodeName, "undo")
			} else {
				// point to all of undone jobs of a sort of type job  in worker servers
				undone = Gcounter.Get(p.NodeName, e.Tags["type"])

			}

			processors[k].Undone = undone

			// if num < MaxExecutionLimit {
			// 	newProcessors = append(newProcessors, p)
			// }
		}

		sort.Sort(ProcessorList(processors))

		log.WithFields(log.Fields{
			"processors":  processors,
			"jobTagsType": e.Tags["type"],
			"minimum":     processors[0].Undone,
		}).Debug("CheckCounter: sort processor by undone.")

	}

	return processors
}

// ExecList stores a slice of Executions.
// This slice can be sorted to provide a time ordered slice of Executions.
type ExecList []*Execution

func (el ExecList) Len() int {
	return len(el)
}

func (el ExecList) Swap(i, j int) {
	el[i], el[j] = el[j], el[i]
}

func (el ExecList) Less(i, j int) bool {
	return el[i].StartedAt.Before(el[j].StartedAt)
}
