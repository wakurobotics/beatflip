package supervisor

import (
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/spf13/viper"
)

func TestSupervise(t *testing.T) {
	viper.Set("service.bin", "sleep")
	viper.Set("service.args", "5")
	viper.Set("updater.enabled", false)

	go func() {
		time.Sleep(1 * time.Second)
		p, err := os.FindProcess(os.Getpid())
		if err != nil {
			t.Errorf("os.FindProcess() returned an error: %v", err)
		}
		p.Signal(syscall.SIGTERM)
	}()

	err := Supervise()
	if err != nil {
		t.Errorf("Supervise() returned an error: %v", err)
	}
}
