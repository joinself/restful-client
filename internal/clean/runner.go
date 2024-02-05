package clean

import (
	"github.com/jasonlvhit/gocron"
	_ "github.com/lib/pq"
)

type Runner interface {
	Run()
}

type RunnerConfig struct {
	Service Service
}

type runner struct {
	service Service
}

func NewRunner(config RunnerConfig) Runner {
	return &runner{config.Service}
}

func (r *runner) Run() {
	gocron.Every(1).Day().Do(r.service.Clean)
	<-gocron.Start()
}
