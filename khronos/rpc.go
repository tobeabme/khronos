package khronos

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/smallnest/rpcx/client"
	"github.com/smallnest/rpcx/server"
)

type RPCServer struct {
	agent *Agent
}

type RPCClient struct {
	//Addres of the server to call
	ServerAddr []*Processor
	xclient    client.XClient
	agent      *Agent
}

type RPCReply struct {
	Success bool
	Ack     int
}

func listenRPC(a *Agent) {
	//as a server
	r := &RPCServer{
		agent: a,
	}
	addr := fmt.Sprintf("%s:%d", a.config.BindIP, a.config.RPCPort)
	s := server.NewServer()
	s.RegisterName("khronos", r, "")
	s.Serve("tcp", addr)
}

// it means that the client is down when servers accept a ServNodeReg request.
func (r *RPCServer) ServNodeReg(ctx context.Context, args *Processor, reply *RPCReply) error {
	// need to remove unfinished executions
	err := r.agent.store.DeleteExecutionsByNodeName(args.NodeName)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("remove unfinished executions before ServNodeReg.SetProcessor")
	}

	// a maxinum number that server can do
	if args.MaxExecutionLimit == 0 {
		args.MaxExecutionLimit = MaxExecutionLimit
	}

	err = r.agent.store.SetProcessor(args)
	if err != nil {
		log.WithFields(log.Fields{
			"processor": args,
		}).Error("RPCServer: ServNodeReg failed.")
	} else {
		reply.Ack = reply.Ack + 1
		reply.Success = true
	}

	return err
}

func (r *RPCServer) MakeJob(ctx context.Context, args *Job, reply *RPCReply) error {

	err := r.agent.store.SetJob(args)
	if err != nil {
		log.WithFields(log.Fields{
			"job": args,
		}).Error("RPCServer: MakeJob failed.")
	} else {
		reply.Ack = reply.Ack + 1
		reply.Success = true
	}

	return err
}

func (r *RPCServer) ExecutionDone(ctx context.Context, args *Execution, reply *RPCReply) error {
	log.WithFields(log.Fields{
		"execution": args,
		"reply":     reply,
	}).Debug("RPCServer: ExecutionDone be called by workerRPC.ExecutionDone.")

	args.FinishedAt = time.Now()
	_, err := r.agent.store.SetExecution(args)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("RPCServer: SetExecution fail.")
	}

	job, err := r.agent.store.GetJob(args.JobName)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("RPCServer: GetJob fail.")
	}

	if args.Success == true {
		job.Metadata.SuccessCount += 1
		job.Metadata.LastSuccess = time.Now()

	} else {
		job.Metadata.ErrorCount += 1
		job.Metadata.LastError = time.Now()
	}

	err = r.agent.store.SetJob(job)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("RPCServer: SetJob fail.")
	}

	reply.Ack = reply.Ack + 1
	reply.Success = true
	return nil
}

func (rc *RPCClient) ExecutionDo(args *Execution) {
	for _, p := range rc.ServerAddr {
		addr := fmt.Sprintf("tcp@%s:%d", p.IP, p.Port)

		d := client.NewPeer2PeerDiscovery(addr, "")
		rc.xclient = client.NewXClient("Worker", client.Failtry, client.RandomSelect, d, client.DefaultOption)
		defer rc.xclient.Close()

		args.NodeName = p.NodeName
		log.WithFields(log.Fields{
			"Node":      p,
			"Execution": args,
		}).Debug("ExecutionDo assign job to work node")

		rpcReply := &RPCReply{}
		err := rc.xclient.Call(context.Background(), "ExecutionDo", args, rpcReply)
		if err != nil {
			log.WithFields(log.Fields{
				"err": err,
			}).Error("failed to call")

		} else {
			log.WithFields(log.Fields{
				"rpcReply":  rpcReply,
				"Execution": args,
			}).Debug("RPCClient: Call Worker.ExecutionDo.")

			if rpcReply.Ack > 0 {
				rc.agent.store.SetExecution(args)
			}

		}
	}

}

//Ping do nothing but just call pong
func (rc *RPCClient) Ping(node *Processor) {
	for {
		addr := fmt.Sprintf("tcp@%s:%d", node.IP, node.Port)
		d := client.NewPeer2PeerDiscovery(addr, "")

		rc.xclient = client.NewXClient("Worker", client.Failtry, client.RandomSelect, d, client.DefaultOption)
		defer rc.xclient.Close()

		log.WithFields(log.Fields{
			"addr": addr,
		}).Debug("PING in a cyclic")

		rpcReply := &RPCReply{}
		err := rc.xclient.Call(context.Background(), "Pong", &struct{}{}, rpcReply)

		log.WithFields(log.Fields{
			"addr":  addr,
			"reply": rpcReply,
		}).Debug("PING reply")

		if err != nil || rpcReply.Success == false {
			log.WithFields(log.Fields{
				"err": err,
			}).Error("PING: failed to call")

			err := rc.agent.store.DeleteExecutionsByNodeName(node.NodeName)
			if err != nil {
				log.WithFields(log.Fields{
					"err": err,
				}).Error("remove unfinished executions after ping failed")
			}

			key := fmt.Sprintf("%s:%d", node.IP, node.Port)
			if _, err := rc.agent.store.DeleteProcessor(node.Application, key); err != nil {
				log.Error("ping error deleting processor: ", err)
			}

			break
		}
		time.Sleep(2 * time.Second)
	}
}

func (rc *RPCClient) SetXClient(ex *Execution) client.XClient {
	addrKVPair := []*client.KVPair{}
	for _, p := range rc.ServerAddr {
		addr := fmt.Sprintf("tcp@%s:%d", p.IP, p.Port)
		fmt.Println("Worker-RPC-Addr: ", addr)
		addrKVPair = append(addrKVPair, &client.KVPair{Key: addr})
	}

	if ex.Concurrency == "allow" {
		d := client.NewMultipleServersDiscovery(addrKVPair)
		rc.xclient = client.NewXClient("Worker", client.Failover, client.RoundRobin, d, client.DefaultOption)

	} else {
		rand.Seed(time.Now().Unix())
		idx := rand.Intn(len(rc.ServerAddr))
		addr := fmt.Sprintf("tcp@%s:%d", rc.ServerAddr[idx].IP, rc.ServerAddr[idx].Port)
		d := client.NewPeer2PeerDiscovery(addr, "")
		rc.xclient = client.NewXClient("Worker", client.Failtry, client.RandomSelect, d, client.DefaultOption)
		ex.NodeName = rc.ServerAddr[idx].IP
	}

	return rc.xclient
}
