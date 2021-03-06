package khronos

import (
	"strings"

	"github.com/robfig/cron"
	log "github.com/sirupsen/logrus"
)

type Scheduler struct {
	Cron    *cron.Cron
	Started bool
	Agent   *Agent `json:"-"`
}

func NewScheduler() *Scheduler {
	c := cron.New()

	return &Scheduler{Cron: c, Started: false}
}

func (s *Scheduler) Start(jobs []*Job) {
	for _, job := range jobs {
		if job.Disabled {
			continue
		}

		if job.IsDone {
			continue
		}

		log.WithFields(log.Fields{
			"job": job.Name,
		}).Debug("scheduler: Adding job to cron")

		job.Agent = s.Agent

		schedule := job.Schedule
		schedule = strings.Trim(schedule, "")
		if schedule == "@oneway" {
			job.Run()
		} else {
			s.Cron.AddJob(job.Schedule, job)
		}

	}
	s.Cron.Start()
	s.Started = true

}

func (s *Scheduler) Stop() {
	if s.Started {
		log.Debug("scheduler: Stopping scheduler")
		s.Cron.Stop()
		s.Started = false
		s.Cron = cron.New()

	}
}

func (s *Scheduler) Restart(jobs []*Job) {
	s.Stop()
	s.Start(jobs)
}
