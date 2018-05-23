package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/smallnest/rpcx/client"
	"github.com/smallnest/rpcx/server"
	"github.com/tobeabme/khronos/khronos"
)

type WorkerRPCServer struct {
	rc *WorkerRPCClient
}

type WorkerRPCClient struct {
	xclient client.XClient
}

var (
	khronosRPCAddr = flag.String("khronosRPCAddr", "tcp@localhost:10005", "khronos rpc server address")
	addr           = flag.String("addr", "127.0.0.1:9002", "worker server address")
)

func main() {
	flag.Parse()

	//set up rpcclient
	xclient := SetXClient()
	defer xclient.Close()

	rc := &WorkerRPCClient{
		xclient: xclient,
	}

	//set up rpcserver
	rs := &WorkerRPCServer{
		rc: rc,
	}

	//start up
	// rc.MakeJob()

	go func() {
		for {
			conn, err := net.Dial("tcp", *addr)
			if err != nil && conn == nil {
				log.Error("net.Dial: ", err, conn)
			} else {
				//after listen running
				rc.ServNodeReg()
				conn.Close()
				return
			}

			time.Sleep(2 * time.Second)
		}
	}()

	rs.listenRPC()

	select {}
}

func (rs *WorkerRPCServer) listenRPC() {
	s := server.NewServer()
	s.RegisterName("Worker", rs, "")
	s.Serve("tcp", *addr)
}

//ExecutionDo start to handle a job
func (rs *WorkerRPCServer) ExecutionDo(ctx context.Context, args *khronos.Execution, reply *khronos.RPCReply) error {
	//start to process this job
	reply.Ack = reply.Ack + 1
	go rs.Process(args)

	return nil
}

//Process is customized
//if finished then recall ExecutionDone
func (rs *WorkerRPCServer) Process(args *khronos.Execution) {
	var done = make(chan bool, 1)
	go func() {
		i := 0
		for {
			i++
			fmt.Println("fmt-Process job:", i)
			log.Debug("log-Process job:", i)
			if i > 5 {
				//if done well put true in done
				done <- true
				break
			}
		}
	}()

	rs.rc.ExecutionDone(args, <-done)
}

//Pong do nothing but just reply to ping
func (rs *WorkerRPCServer) Pong(ctx context.Context, args *struct{}, reply *khronos.RPCReply) error {
	reply.Ack = reply.Ack + 1
	reply.Success = true
	return nil
}

//ServNodeReg offer that register a addr to etcd by rpc
func (rc *WorkerRPCClient) ServNodeReg() {
	servNode := &khronos.Processor{
		Application:       "spider",
		NodeName:          "server-001",
		IP:                "127.0.0.1",
		Port:              9002,
		MaxExecutionLimit: 10,
		Status:            true,
	}

	replay := &khronos.RPCReply{}

	err := rc.xclient.Call(context.Background(), "ServNodeReg", servNode, replay)
	if err != nil {
		log.Error("failed to call: ", err)
	}

	log.Debug("ServNodeReg ack: %d, success: %t", replay.Ack, replay.Success)
}

//MakeJob can produce a lots of job, just over and over again.
func (rc *WorkerRPCClient) MakeJob() {
	testJob := &khronos.Job{
		Name:        "testJob1",
		Breif:       "the description in breif",
		Schedule:    "@every 2s",
		JobType:     "rpc",
		Payload:     map[string]string{"eth": "coinid-1"},
		Disabled:    false,
		Concurrency: "forbid",
		Application: "spider",
	}

	replay := &khronos.RPCReply{}

	err := rc.xclient.Call(context.Background(), "MakeJob", testJob, replay)
	if err != nil {
		log.Error("failed to create a job: ", err)
	}

	log.Debug("MakeJob ack: %d, success: %t", replay.Ack, replay.Success)

}

//ExecutionDone reply the result of handling a job to khronos-rpcserver
func (rc *WorkerRPCClient) ExecutionDone(args *khronos.Execution, done bool) {
	replay := &khronos.RPCReply{}
	args.Success = done

	err := rc.xclient.Call(context.Background(), "ExecutionDone", args, replay)
	if err != nil {
		log.Error("failed to call: ", err)
	}
	fmt.Println("wokr done", replay)
}

func SetXClient() client.XClient {
	d := client.NewMultipleServersDiscovery([]*client.KVPair{{Key: *khronosRPCAddr}})
	xclient := client.NewXClient("khronos", client.Failover, client.RoundRobin, d, client.DefaultOption)

	return xclient
}
