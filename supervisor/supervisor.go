package supervisor

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var startSignal = make(chan struct{}, 1)

func Supervise() error {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	startSignal <- struct{}{}

	ctx, cancelUpdates := context.WithCancel(context.Background())
	if viper.GetBool("updater.enabled") {
		go check_updates(ctx)
	}

	for {
		select {
		case <-startSignal:
			updateWg.Wait()
			go start()
		case sig := <-sigs:
			if sig == syscall.SIGHUP {
				if err := restart(); err != nil {
					logrus.WithError(err).Error("received SIGHUP, but failed to restart service")
				}
				continue
			}

			cancelUpdates()
			if err := terminate(sig); err != nil {
				return fmt.Errorf("failed to terminate service: %+v", err)
			}
			return nil
		}
	}
}
