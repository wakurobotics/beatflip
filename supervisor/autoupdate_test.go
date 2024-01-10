package supervisor

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSupervisor_check_updates(t *testing.T) {
	config := validConfig()
	config.AutoUpdate.Enabled = true

	supervisor, err := NewSupervisor(config)
	assert.Nil(t, err)
	assert.NotNil(t, supervisor)

	running := &atomic.Bool{}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		running.Store(true)
		supervisor.checkUpdates(ctx)
		running.Store(false)
	}()

	time.Sleep(100 * time.Millisecond)
	assert.True(t, running.Load())
	cancel()
	time.Sleep(100 * time.Millisecond)
	assert.False(t, running.Load())
}

func TestSupervisor_getManifest(t *testing.T) {
	want := &manifest{
		Version: "1.0.0",
		Sha256:  "b026324c6904b2a9cb4b88d6d61c81d1000000",
	}
	payload, err := json.Marshal(want)
	assert.Nil(t, err)

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Write(payload)
	}))
	defer server.Close()

	config := validConfig()
	config.AutoUpdate.Enabled = true
	config.AutoUpdate.Server = server.URL

	supervisor, err := NewSupervisor(config)
	assert.Nil(t, err)
	assert.NotNil(t, supervisor)

	got, err := supervisor.getManifest()
	assert.Nil(t, err)
	assert.Equal(t, want, got)
}

func TestSupervisor_getCurrentVersion(t *testing.T) {
	config := validConfig()
	config.AutoUpdate.Version.Bin = "a-nonexistent-binary"

	supervisor, err := NewSupervisor(config)
	assert.Nil(t, err)
	assert.NotNil(t, supervisor)

	_, err = supervisor.getCurrentVersion()
	assert.Error(t, err)
}

func Test_manifest_checksum(t *testing.T) {
	m := &manifest{
		Version: "1.0.0",
		Sha256:  "dffd6021bb2bd5b0af676290809ec3a53191dd81c7f70a4b28688a362182986f", // echo -n "Hello, World!" | sha256sum
	}

	validPayload := []byte("Hello, World!")
	assert.Nil(t, m.checksum(validPayload))

	invalidPayload := []byte("Goodbye, World!")
	assert.Error(t, m.checksum(invalidPayload))
}

func TestSupervisor_download(t *testing.T) {
	want := []byte("Hello, World!")
	m := &manifest{
		Version: "1.0.0",
		Sha256:  "dffd6021bb2bd5b0af676290809ec3a53191dd81c7f70a4b28688a362182986f", // echo -n "Hello, World!" | sha256sum
	}

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		gzipWriter := gzip.NewWriter(rw)
		defer gzipWriter.Close()

		rw.Header().Set("Content-Encoding", "gzip")
		rw.Header().Set("Content-Type", "application/octet-stream")

		_, err := gzipWriter.Write(want)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	}))
	defer server.Close()

	config := validConfig()
	config.AutoUpdate.Enabled = true
	config.AutoUpdate.Server = server.URL

	supervisor, err := NewSupervisor(config)
	assert.Nil(t, err)
	assert.NotNil(t, supervisor)

	got, err := supervisor.download(m)
	assert.Nil(t, err)
	assert.Equal(t, want, got)
}
