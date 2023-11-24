package supervisor

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"syscall"

	"github.com/sirupsen/logrus"
)

var empty = struct{}{}

// ServiceConfig represents the configuration of a service.
type ServiceConfig struct {
	// Command represents the command to be executed.
	Command `mapstructure:",squash"`

	// Instances represents the number of instances to be created.
	Instances int `mapstructure:"instances"`

	// AutoUpdate represents the auto-update configuration.
	AutoUpdate *AutoUpdate `mapstructure:"autoupdate"`
}

func (c *ServiceConfig) validate() error {
	if err := c.Command.validate(); err != nil {
		return fmt.Errorf("invalid command: %w", err)
	}

	if err := c.AutoUpdate.validate(); err != nil {
		return err
	}

	if c.Instances < 0 {
		c.Instances = 0
	}

	return nil
}

// Supervisor represents the supervisor of a service.
type Supervisor struct {
	config *ServiceConfig

	instances map[*exec.Cmd]struct{}
	mu        *sync.Mutex

	startSignal chan struct{}
	updateWg    sync.WaitGroup

	osSignals chan os.Signal
}

// NewSupervisor creates a new supervisor.
func NewSupervisor(config *ServiceConfig) (*Supervisor, error) {
	if err := config.validate(); err != nil {
		return nil, fmt.Errorf("invalid service config: %w", err)
	}

	return &Supervisor{
		config: config,

		instances: make(map[*exec.Cmd]struct{}),
		mu:        &sync.Mutex{},

		osSignals: make(chan os.Signal, 1),

		startSignal: make(chan struct{}, config.Instances),
		updateWg:    sync.WaitGroup{},
	}, nil
}

// Run runs the supervisor.
func (s *Supervisor) Run() error {
	signal.Notify(s.osSignals, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	ctx, cancelUpdates := context.WithCancel(context.Background())
	if s.config.AutoUpdate.Enabled {
		go s.check_updates(ctx)
	}

	s.boot()

	for {
		select {
		case <-s.startSignal:
			s.updateWg.Wait()
			go s.spawnInstance()
		case sig := <-s.osSignals:
			if sig == syscall.SIGHUP {
				if err := s.terminateAll(syscall.SIGTERM); err != nil {
					logrus.WithError(err).Error("received SIGHUP, but failed to restart service")
				}
				continue
			}

			cancelUpdates()
			if err := s.terminateAll(sig); err != nil {
				return fmt.Errorf("failed to terminate service: %+v", err)
			}
			return nil
		}
	}
}

func (s *Supervisor) boot() {
	for i := 0; i < s.config.Instances; i++ {
		s.startSignal <- empty
	}
}
