package supervisor

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	platform = runtime.GOOS + "-" + runtime.GOARCH

	defaultUpdateInterval = time.Minute * 15
)

type AutoUpdate struct {
	Enabled  bool          `mapstructure:"enabled"`
	Interval time.Duration `mapstructure:"interval"`
	Version  *Command      `mapstructure:"version"`
	Server   string        `mapstructure:"server"`
}

func (a *AutoUpdate) validate() error {
	if !a.Enabled {
		return nil
	}

	if a.Interval <= 0 {
		a.Interval = defaultUpdateInterval
	}

	if a.Server == "" {
		return errors.New("'server' must not be empty when autoupdate is enabled")
	}
	a.Server = strings.TrimRight(a.Server, "/") + "/"

	if _, err := url.Parse(a.Server); err != nil {
		return fmt.Errorf("invalid server URL: %w", err)
	}

	if err := a.Version.validate(); err != nil {
		return fmt.Errorf("invalid version command: %w", err)
	}

	return nil
}

func (s *Supervisor) checkUpdates(ctx context.Context) {
	if err := s.update(); err != nil {
		log.Printf("failed to update: %+v\n", err)
	}

	t := time.NewTicker(s.config.AutoUpdate.Interval)
	defer t.Stop()

	for {
		select {
		case <-t.C:
			if err := s.update(); err != nil {
				log.Printf("failed to update: %+v\n", err)
			}
		case <-ctx.Done():
			return
		}
	}
}

func (s *Supervisor) update() error {
	m, err := s.getManifest()
	if err != nil {
		return err
	}

	currentVersion, err := s.getCurrentVersion()
	if err != nil {
		return err
	}

	manifestVersion, err := m.parseVersion()
	if err != nil {
		return fmt.Errorf("failed to parse manifest Version: %s: %+v", m.Version, err)
	}

	log := logrus.WithFields(logrus.Fields{
		"current": currentVersion,
		"latest":  manifestVersion,
	})

	if compareSemver(currentVersion, manifestVersion) < 1 {
		// version greater or equal; nothing to do
		log.Info("current version greater or equal; nothing to do")
		return nil
	}

	log.Info("a newer version is available; starting download...")
	b, err := s.download(m)
	if err != nil {
		return err
	}

	// start the update process
	s.updateWg.Add(1)
	defer s.updateWg.Done()

	if err := s.terminateAll(syscall.SIGTERM); err != nil {
		return fmt.Errorf("failed to terminate: %+v", err)
	}

	log.Infof("flipping binaries (%v bytes)", len(b))
	if err := os.WriteFile(s.config.Bin, b, 0644); err != nil {
		return err
	}

	// TODO: add post-update hooks here

	return nil
}

func (s *Supervisor) getCurrentVersion() ([]int, error) {
	v, err := s.config.AutoUpdate.Version.cmd().Output()
	if err != nil {
		return nil, err
	}
	return parseSemver(string(v))
}

type manifest struct {
	Version string `json:"version"`
	Sha256  string `json:"sha256"`
}

func (m *manifest) parseVersion() ([]int, error) {
	return parseSemver(m.Version)
}

func (m *manifest) checksum(b []byte) error {
	hash := sha256.Sum256(b)
	sum := hex.EncodeToString(hash[:])
	if sum != m.Sha256 {
		return errors.New("checksum failed")
	}
	return nil
}

func (s *Supervisor) download(m *manifest) ([]byte, error) {
	u := s.config.AutoUpdate.Server + url.QueryEscape(m.Version) + "/" + url.QueryEscape(platform)
	logrus.Infof("downloading newer version from %s", u)
	resp, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if err := m.checksum(b); err != nil {
		return nil, err
	}
	return b, nil
}

func (s *Supervisor) getManifest() (*manifest, error) {
	u := s.config.AutoUpdate.Server + url.QueryEscape(platform) + ".json"
	logrus.Infof("fetching manifest from %s", u)
	resp, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	m := &manifest{}
	if err := json.Unmarshal(b, m); err != nil {
		return nil, err
	}
	return m, nil
}
