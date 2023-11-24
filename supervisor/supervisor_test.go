package supervisor

import (
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/atomic"
)

func validConfig() *ServiceConfig {
	return &ServiceConfig{
		Command: Command{
			Bin:  "sleep",
			Args: []string{"10"},
		},
		Instances: 1,
		AutoUpdate: &AutoUpdate{
			Enabled:  false,
			Interval: 0,
			Version: &Command{
				Bin:  "sleep",
				Args: []string{"--version"},
			},
			Server: "http://localhost:8080/",
		},
	}
}

func TestNewSupervisor(t *testing.T) {
	want := &ServiceConfig{
		Instances: 0,
		AutoUpdate: &AutoUpdate{
			Server:   "http://localhost:8080/",
			Interval: defaultUpdateInterval,
		},
	}
	supervisor, err := NewSupervisor(&ServiceConfig{
		Command: Command{
			Bin: "sleep",
		},
		Instances: -1,
		AutoUpdate: &AutoUpdate{
			Enabled: true,
			Version: &Command{
				Bin: "sleep",
			},
			Server: "http://localhost:8080",
		},
	})

	assert.Nil(t, err)
	assert.NotNil(t, supervisor)

	got := supervisor.config
	assert.Equal(t, want.Instances, got.Instances)
	assert.Equal(t, want.AutoUpdate.Interval, got.AutoUpdate.Interval)
	assert.Equal(t, want.AutoUpdate.Server, got.AutoUpdate.Server)

}

func TestSupervisor_Run(t *testing.T) {
	supervisor, err := NewSupervisor(validConfig())
	assert.Nil(t, err)
	assert.NotNil(t, supervisor)

	running := atomic.NewBool(true)

	go func() {
		err := supervisor.Run()
		running.Store(false)
		assert.Nil(t, err)
	}()
	time.Sleep(100 * time.Millisecond)
	assert.True(t, running.Load())

	supervisor.osSignals <- syscall.SIGHUP
	time.Sleep(100 * time.Millisecond)
	assert.True(t, running.Load())

	supervisor.osSignals <- syscall.SIGTERM
	time.Sleep(100 * time.Millisecond)
	assert.False(t, running.Load())
}

func TestSupervisor_boot(t *testing.T) {
	want := 13
	config := validConfig()
	config.Instances = want

	supervisor, err := NewSupervisor(config)
	assert.Nil(t, err)
	assert.NotNil(t, supervisor)

	supervisor.boot()

	assert.Equal(t, want, len(supervisor.startSignal))
}
