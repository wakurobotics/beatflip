package supervisor

import (
	"errors"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/sirupsen/logrus"
)

type Command struct {
	Bin  string   `mapstructure:"bin"`
	Args []string `mapstructure:"args"`
}

func (c *Command) validate() error {
	if c.Bin == "" {
		return errors.New("bin must not be empty")
	}
	return nil
}

func (c *Command) cmd() *exec.Cmd {
	return exec.Command(c.Bin, c.Args...)
}
func (c *Command) String() string {
	return c.Bin + " " + strings.Join(c.Args, " ")
}

func (s *Supervisor) terminateAll(sig os.Signal) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var errs = []error{}
	for cmd := range s.instances {
		if cmd != nil && cmd.Process != nil && cmd.ProcessState == nil {
			errs = append(errs, cmd.Process.Signal(sig))
		}
	}
	return errors.Join(errs...)
}

func (s *Supervisor) spawnInstance() error {
	cmd := s.config.Command.cmd()
	s.mu.Lock()
	s.instances[cmd] = empty
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		delete(s.instances, cmd)
		s.startSignal <- empty
	}()

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

	log := logrus.WithField("command", s.config.Command.String())
	log.Info("starting service")
	err = cmd.Run()
	log.WithError(err).Info("service stopped")
	return err
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
