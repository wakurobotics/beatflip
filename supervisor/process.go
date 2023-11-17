package supervisor

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var cmd *exec.Cmd

func terminate(sig os.Signal) error {
	if cmd != nil && cmd.Process != nil && cmd.ProcessState == nil {
		return cmd.Process.Signal(sig)
	}
	return nil
}

func start() error {
	defer func() {
		startSignal <- struct{}{}
	}()
	c := viper.GetString("service.bin")
	args := viper.GetStringSlice("service.args")
	cmd = exec.Command(c, args...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	go pipeUntilClosed(stdout, os.Stdout)
	go pipeUntilClosed(stderr, os.Stderr)

	fullCmd := c + " " + strings.Join(args, " ")
	logrus.WithField("command", fullCmd).Info("(re)-starting service")
	err = cmd.Start()
	logrus.WithError(err).Info("service stopped")
	return err
}

func restart() error {
	if err := terminate(syscall.SIGTERM); err != nil {
		return fmt.Errorf("failed to terminate service before restart: %+v", err)
	}
	startSignal <- struct{}{}
	return nil
}

func pipeUntilClosed(r io.ReadCloser, out *os.File) {
	for {
		b, err := io.ReadAll(r)
		if err != nil {
			return
		}
		if _, err := out.Write(b); err != nil {
			return
		}
	}
}
