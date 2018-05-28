package khronos

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"time"

	log "github.com/sirupsen/logrus"
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
	InitLogger(config.LogLevel, config.LogPath)
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

	go func() {
		for {
			rpcSrvAddr := fmt.Sprintf("%s:%d", a.config.BindIP, a.config.RPCPort)
			conn, err := net.Dial("tcp", rpcSrvAddr)
			if err != nil && conn == nil {
				fmt.Println("net.Dial: ", rpcSrvAddr, err, conn)
			} else {
				go a.HeartBeat()
				go a.Schedule()
				conn.Close()
				return
			}

			time.Sleep(2 * time.Second)
		}
	}()

	listenRPC(a)
}

//Schedule is reponsible for adding job to cron.
//Start or restart scheduler
func (a *Agent) Schedule() {
	for {
		log.Debug("agent.schedule has been invoked.")
		jobs, err := a.store.GetJobs()
		if err != nil {
			log.Error("agent.schedule", err)
		}
		a.sched.Restart(jobs)
		time.Sleep(15 * time.Minute)
	}

	//===========================================
	//expect only watch create event,but put has both event add and update
	// events, err := a.store.WatchJobsTree()
	// log.WithFields(log.Fields{
	// 	"err":    err,
	// 	"events": events,
	// }).Debug("Watch Jobs")
	// start := 0

	// for {
	// 	select {
	// 	case pairs := <-events:
	// 		start++
	// 		// Do something with events
	// 		var jobs []*Job
	// 		var onewayJobs []*Job
	// 		for _, pair := range pairs {
	// 			job := &Job{}
	// 			//skipping del event
	// 			if len(pair.Value) == 0 {
	// 				continue
	// 			}
	// 			if err := json.Unmarshal([]byte(string(pair.Value)), &job); err != nil {
	// 				log.Error(err)
	// 			} else {
	// 				schedule := job.Schedule
	// 				schedule = strings.Trim(schedule, "")
	// 				if schedule == "@oneway" {
	// 					onewayJobs = append(onewayJobs, job)
	// 				} else {
	// 					jobs = append(jobs, job)
	// 				}

	// 			}

	// 		}
	// 		if start == 1 {
	// 			log.WithFields(log.Fields{
	// 				"start": start,
	// 			}).Debug("sched restart")
	// 			a.sched.Restart(jobs)
	// 		} else {
	// 			log.WithFields(log.Fields{
	// 				"start": start,
	// 			}).Debug("sched start")
	// 			a.sched.Start(jobs)
	// 		}

	// 	}
	// }
}

//HeartBeat detect work node
func (a *Agent) HeartBeat() {
	rc := &RPCClient{
		agent: a,
	}

	events, err := a.store.WatchProcessorTree()
	log.WithFields(log.Fields{
		"err":    err,
		"events": events,
	}).Debug("Watch Processor")

	for {
		select {
		case pairs := <-events:
			// Do something with events
			for _, pair := range pairs {
				node := &Processor{}
				//skipping del event
				if len(pair.Value) == 0 {
					continue
				}
				if err := json.Unmarshal([]byte(string(pair.Value)), &node); err != nil {
					log.Error(err)
				} else {
					log.WithFields(log.Fields{
						"node": node,
					}).Debug("HeartBeat.events")

					time.Sleep(2 * time.Second)
					go rc.Ping(node)
				}

			}
		}
	}

}

func (a *Agent) Do(ex *Execution) {
	log.WithFields(log.Fields{
		"ex": ex,
	}).Debug("agent.Do has been trigger.")

	srvAddr := a.GetWorkerRPCAddr(ex, "")
	log.WithFields(log.Fields{
		"ex":      ex,
		"srvAddr": srvAddr,
	}).Debug("agent.Do invoked agent.GetWorkerRPCAddr to get worker nodes.")

	if srvAddr == nil {
		log.WithFields(log.Fields{
			"ex":      ex,
			"srvAddr": srvAddr,
		}).Error("agent.Do Not found any worker node.")

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

func (a *Agent) GetWorkerRPCAddr(ex *Execution, rebalance string) []*Processor {
	log.WithFields(log.Fields{
		"ex": ex,
	}).Debug("agent.getWorkerRPCAddr has been called.")

	srvAddr, err := a.store.GetProcessorsByApp(ex.Application)
	if err != nil {
		log.WithFields(log.Fields{
			"Application": ex.Application,
			"err":         err,
		}).Error("agent.getWorkerRPCAddr don't got any processor.")
		return nil
	}

	srvAddr = ex.CheckCounter(srvAddr)

	if ex.Concurrency == "forbid" && len(srvAddr) > 0 {

		switch rebalance {
		case "random":
			rand.Seed(time.Now().Unix())
			idx := rand.Intn(len(srvAddr))
			srvAddrRand := make([]*Processor, 0)
			srvAddrRand = append(srvAddrRand, srvAddr[idx])

			log.WithFields(log.Fields{
				"srvAddr": srvAddr[idx],
			}).Debug("agent.GetWorkerRPCAddr random selection")

			return srvAddrRand

		// return a processor that being the least amount of undo of processor
		default:
			idx := 0
			srvAddrDef := make([]*Processor, 0)
			srvAddrDef = append(srvAddrDef, srvAddr[idx])

			log.WithFields(log.Fields{
				"srvAddr": srvAddr[idx],
			}).Debug("agent.GetWorkerRPCAddr minimum selection by undone")

			return srvAddrDef

		}

	}

	return srvAddr
}

func (a *Agent) Leave() error {
	return nil
}
