package khronos

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/abronan/valkeyrie/store"
)

var (
	logLevel = "error"
	etcdAddr = getEnvWithDefault()
)

func getEnvWithDefault() string {
	ea := os.Getenv("khronos_BACKEND_MACHINE")
	if ea == "" {
		return "127.0.0.1:2379"
	}
	return ea
}

//go test -v -run=TestStoreJob
func TestStoreJob(t *testing.T) {
	store := createTestStore()

	// Cleanup everything
	if err := cleanTestKVSpace(store); err != nil {
		t.Logf("error cleaning up: %v", err)
	}

	testJob := &Job{
		Name:        "test3",
		Schedule:    "@every 5s",
		JobType:     "rpc",
		Payload:     map[string]string{"eos": "coinid-3"},
		Disabled:    false,
		Concurrency: "forbid",
		Application: "spider",
		Tags:        map[string]string{"type": "websocket"},
	}

	// Check that we still get an empty job list
	jobs, err := store.GetJobs()
	if err != nil {
		t.Fatalf("error getting jobs: %s", err)
	} else if jobs == nil {
		t.Fatal("jobs empty, expecting empty slice")
	}

	if err := store.SetJob(testJob); err != nil {
		t.Fatalf("error creating job: %s", err)
	}

	jobs, err = store.GetJobs()
	if err != nil {
		t.Fatalf("error getting jobs: %s", err)
	}
	if jobs[0].Name != "test" {
		t.Fatalf("expected job name: %s got: %s", testJob.Name, jobs[0].Name)
	}
	fmt.Println("Got all jobs in test", jobs)

	if _, err := store.DeleteJob("test"); err != nil {
		t.Fatalf("error deleting job: %s", err)
	}

	if _, err := store.DeleteJob("test"); err != nil {
		t.Fatalf("error job deletion should fail: %s", err)
	}

}

//go test -v -run=TestStoreProcessor
func TestStoreProcessor(t *testing.T) {
	store := createTestStore()

	// Cleanup everything
	if err := cleanTestKVSpace(store); err != nil {
		t.Logf("error cleaning up: %v", err)
	}

	testProcessor := &Processor{
		Application: "spider",
		NodeName:    "server-009",
		IP:          "127.0.0.1",
		Port:        9009,
		Status:      true,
	}

	// Check that we still get an empty processor list
	processors, err := store.GetProcessors()
	if err != nil {
		t.Fatalf("error getting processors: %s", err)
	} else if processors == nil {
		t.Fatal("processors empty, expecting empty slice")
	}

	if err := store.SetProcessor(testProcessor); err != nil {
		t.Fatalf("error creating processor: %s", err)
	}

	processors, err = store.GetProcessors()
	if err != nil {
		t.Fatalf("error getting processors: %s", err)
	}
	if processors[0].NodeName != "server-001" {
		t.Fatalf("expected node name: %s got: %s", testProcessor.NodeName, processors[0].NodeName)
	}
	fmt.Println("Got all processors in test", processors)

	addr := fmt.Sprintf("%s:%d", testProcessor.IP, testProcessor.Port)
	if _, err := store.DeleteProcessor(testProcessor.Application, addr); err != nil {
		t.Fatalf("error deleting processor: %s", err)
	}

	if _, err := store.DeleteProcessor(testProcessor.Application, addr); err != nil {
		t.Fatalf("error processor deletion should fail: %s", err)
	}

}

//go test -v -run=TestStoreExecution
func TestStoreExecution(t *testing.T) {
	s := createTestStore()

	ex1 := &Execution{
		JobName:     "test1",
		StartedAt:   time.Now(),
		Success:     false,
		NodeName:    "server-001",
		Application: "spider",
	}
	ex2 := &Execution{
		JobName:     "test1",
		StartedAt:   ex1.StartedAt,
		FinishedAt:  time.Now(),
		Success:     true,
		NodeName:    "server-001",
		Application: "spider",
	}
	_, err := s.SetExecution(ex1)
	if err != nil {
		t.Fatalf("error creating execution1: %s", err)
	}
	time.Sleep(60)
	_, err = s.SetExecution(ex2)
	if err != nil {
		t.Fatalf("error creating execution2: %s", err)
	}
}

func createTestStore() *Store {
	store := NewStore("etcdv3", []string{"127.0.0.1:2379"}, "/khronos-test")
	return store
}

func cleanTestKVSpace(s *Store) error {
	err := s.Client.DeleteTree("/khronos-test")
	if err != nil && err != store.ErrKeyNotFound {
		return err
	}
	return nil
}
