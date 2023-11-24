package cmd

import (
	"errors"
	"fmt"
	"sync"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/wakurobotics/beatflip/supervisor"
)

var superviseCmd = &cobra.Command{
	Use: "supervise",
	RunE: func(cmd *cobra.Command, args []string) error {
		m := &sync.Mutex{}
		errs := []error{}

		wg := sync.WaitGroup{}
		for name := range viper.GetStringMap("services") {
			c := &supervisor.ServiceConfig{}
			if err := viper.UnmarshalKey("services."+name, c); err != nil {
				return err
			}
			s, err := supervisor.NewSupervisor(c)
			if err != nil {
				return fmt.Errorf("failed to create supervisor for service %s: %w", name, err)
			}

			wg.Add(1)

			go func(s *supervisor.Supervisor) {
				err := s.Run()
				m.Lock()
				defer m.Unlock()
				errs = append(errs, err)
				wg.Done()
			}(s)
		}

		wg.Wait()

		return errors.Join(errs...)
	},
}

func init() {
	rootCmd.AddCommand(superviseCmd)
}
