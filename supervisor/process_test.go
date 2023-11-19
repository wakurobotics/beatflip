package supervisor

import (
	"os"
	"testing"
	"time"

	"github.com/spf13/viper"
)

func TestStart(t *testing.T) {
	viper.Set("service.bin", "sleep")
	viper.Set("service.args", []string{"15"})

	go func() {
		if err := start(); err != nil {
			t.Errorf("start() returned an error: %v", err)
		}
	}()
	time.Sleep(1 * time.Second)

	if cmd == nil || cmd.Process == nil || cmd.ProcessState != nil || cmd.Process.Pid <= 0 {
		t.Errorf("start() failed to start a process")
	}
}

func TestStartAndTerminate(t *testing.T) {
	viper.Set("service.bin", "sleep")
	viper.Set("service.args", []string{"3"})

	go func() {
		if err := start(); err != nil {
			t.Errorf("start() returned an error: %v", err)
		}
	}()
	time.Sleep(1 * time.Second)

	if err := terminate(os.Interrupt); err != nil {
		t.Errorf("terminate() returned an error: %v", err)
	}
}

func TestRestart(t *testing.T) {
	viper.Set("service.bin", "sleep")
	viper.Set("service.args", []string{"3"})

	go func() {
		if err := start(); err != nil {
			t.Errorf("start() returned an error: %v", err)
		}
	}()
	time.Sleep(1 * time.Second)

	go func() {
		for {
			<-startSignal
		}
	}()
	if err := restart(); err != nil {
		t.Errorf("restart() returned an error: %v", err)
	}

}
