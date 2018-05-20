package khronos

import (
	"fmt"
	"testing"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/robfig/cron"
)

// Many tests schedule a job for every second, and then wait at most a second
// for it to run.  This amount is just slightly larger than 1 second to
// compensate for a few milliseconds of runtime.
const OneSecond = 1*time.Second + 10*time.Millisecond

type DummyJob struct{}

func (d DummyJob) Run() {
	fmt.Println("YOLO")
	log.Info("YOLO....")
}

func TestCronAddJob(t *testing.T) {
	var job DummyJob

	c := cron.New()
	c.Start()
	defer c.Stop()
	c.AddJob("* * * * * ?", job)

	select {
	case <-time.After(OneSecond):
		return
	}
}
