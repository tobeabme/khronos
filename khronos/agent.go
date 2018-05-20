package khronos

import (
	"math/rand"
	"time"

	log "github.com/Sirupsen/logrus"
)

type Agent struct {
	store      *Store
	sched      *Scheduler
	config     *Configuration
	ShutdownCh <-chan struct{}
}

// The returned value is the exit code.
func (a *Agent) Run() {
	log.Debug("agent.run has been called...")
	a.config = NewConfig()
	a.Start()
}

func NewAgent(config *Configuration) *Agent {
	a := &Agent{config: config}
	return a
}

func (a *Agent) Start() error {
	a.StartServer()
	return nil
}

// startServer handles all necessary startup functions for a server
func (a *Agent) StartServer() {
	log.Debug("agent.StartServer has been called...")
	a.store = NewStore(a.config.Backend, a.config.BackendMachines, a.config.Keyspace)
	a.sched = NewScheduler()
	a.sched.Agent = a
	a.Schedule()
	a.HeartBeat()
	listenRPC(a)
}

//Schedule is reponsible for adding job to cron.
//Start or restart scheduler
func (a *Agent) Schedule() {
	log.Debug("agent.schedule has been called...")
	jobs, err := a.store.GetJobs()
	if err != nil {
		log.Error("agent.schedule", err)
	}
	a.sched.Restart(jobs)
}

//HeartBeat detect work node
func (a *Agent) HeartBeat() {
	rc := &RPCClient{
		agent: a,
	}
	nodes, err := a.store.GetProcessors()
	if err == nil {
		for _, node := range nodes {
			go rc.Ping(node)
		}
	}

}

func (a *Agent) Do(ex *Execution) {
	log.WithFields(log.Fields{
		"ex": ex,
	}).Debug("agent.Do has been trigger.")

	srvAddr := a.GetWorkerRPCAddr(ex)
	log.WithFields(log.Fields{
		"ex":      ex,
		"srvAddr": srvAddr,
	}).Debug("agent.exec invoked agent.GetWorkerRPCAddr to get worker nodes.")

	if srvAddr == nil {
		log.WithFields(log.Fields{
			"ex":      ex,
			"srvAddr": srvAddr,
		}).Error("agent.exec Not found any worker node.")

	} else {
		rc := &RPCClient{
			ServerAddr: srvAddr,
			agent:      a,
		}

		// rc.SetXClient(ex)
		// defer rc.xclient.Close()

		rc.ExecutionDo(ex)
	}

}

func (a *Agent) GetWorkerRPCAddr(ex *Execution) []*Processor {
	log.WithFields(log.Fields{
		"ex": ex,
	}).Debug("agent.getWorkerRPCAddr has been called.")

	srvAddr, err := a.store.GetProcessorsByApp(ex.Application)
	if err != nil {
		log.WithFields(log.Fields{
			"Application": ex.Application,
			"err":         err,
		}).Error("agent.getWorkerRPCAddr called store.GetProcessorsByApp.")
		return nil
	}

	if ex.Concurrency == "forbid" {
		rand.Seed(time.Now().Unix())
		idx := rand.Intn(len(srvAddr))
		srvAddrRand := make([]*Processor, 0)
		srvAddrRand = append(srvAddrRand, srvAddr[idx])
		return srvAddrRand
	}

	return srvAddr
}

func (a *Agent) Leave() error {
	return nil
}
