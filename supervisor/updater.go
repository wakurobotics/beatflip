package supervisor

import (
	"bytes"
	"compress/gzip"
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
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"golang.org/x/mod/semver"
)

const (
	platform = runtime.GOOS + "-" + runtime.GOARCH

	defaultUpdateInterval = time.Minute * 15
)

var updateWg sync.WaitGroup = sync.WaitGroup{}

func check_updates(ctx context.Context) {
	updateInterval := viper.GetDuration("updater.interval")
	if updateInterval <= 0 {
		updateInterval = defaultUpdateInterval
	}

	t := time.NewTicker(updateInterval)
	defer t.Stop()

	for {
		select {
		case <-t.C:
			if err := update(); err != nil {
				log.Printf("failed to update: %+v\n", err)
			}
		case <-ctx.Done():
			return
		}
	}
}

func update() error {
	url := viper.GetString("updater.url")
	m, err := getManifest(url)
	if err != nil {
		return err
	}

	v, err := getCurrentVersion()
	if err != nil {
		return err
	}
	currentVersion := string(v)

	if false {
		if semver.Compare(m.Version, currentVersion) < 1 {
			// version greater or equal; nothing to do
			logrus.WithFields(logrus.Fields{
				"current": currentVersion,
				"latest":  m.Version,
			}).Info("current version greater or equal; nothing to do")
			return nil
		}
	}
	logrus.WithFields(logrus.Fields{
		"current": currentVersion,
		"latest":  m.Version,
	}).Info("a newer version is available; starting download...")

	b, err := m.download()
	if err != nil {
		return err
	}

	// start the update process
	updateWg.Add(1)
	defer updateWg.Done()

	if err := terminate(syscall.SIGTERM); err != nil {
		return fmt.Errorf("failed to terminate: %+v", err)
	}

	// TODO: impl
	fmt.Printf("TODO: flipping binaries (%v bytes)\n", len(b))

	return nil
}

func getCurrentVersion() ([]byte, error) {
	c := viper.GetString("updater.version.bin")
	args := viper.GetStringSlice("updater.version.args")
	return exec.Command(c, args...).Output()
}

type manifest struct {
	base    string
	Version string
	Sha256  string
}

func (m *manifest) checksum(b []byte) error {
	hash := sha256.Sum256(b)
	sum := hex.EncodeToString(hash[:])
	if sum != m.Sha256 {
		return errors.New("checksum failed")
	}
	return nil
}

func (m *manifest) download() ([]byte, error) {
	u := strings.TrimRight(m.base, "/") + "/" + url.QueryEscape(m.Version) + "/" + url.QueryEscape(platform) + ".gz"
	logrus.Infof("downloading newer version from %s", u)
	resp, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if err := m.checksum(b); err != nil {
		return nil, err
	}
	r := bytes.NewReader(b)
	gz, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}
	buf := new(bytes.Buffer)
	if _, err = io.Copy(buf, gz); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func getManifest(base string) (*manifest, error) {
	u := strings.TrimRight(base, "/") + "/" + url.QueryEscape(platform) + ".json"
	logrus.Infof("fetching manifest from %s", u)
	resp, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	m := &manifest{base: base}
	if err := json.Unmarshal(b, m); err != nil {
		return nil, err
	}
	return m, nil
}
