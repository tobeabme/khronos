package khronos

import (
	"testing"
	"time"
)

//go test -v -run=TestSchedule
func TestSchedule(t *testing.T) {
	sched := NewScheduler()

	testJob1 := &Job{
		Name:       "cron_job",
		Schedule:   "@every 2s", //* * * * * ?
		Command:    "echo 'test1'",
		Owner:      "John Dough",
		OwnerEmail: "foo@bar.com",
		Shell:      true,
		JobType:    "shell",
		Disabled:   false,
	}
	sched.Start([]*Job{testJob1})

	select {
	case <-time.After(10 * time.Second):
		return
	}

}
