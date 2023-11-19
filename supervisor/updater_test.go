package supervisor

import (
	"bytes"
	"compress/gzip"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetManifest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Write([]byte(`{"version": "1.0.0", "sha256": "b026324c6904b2a9cb4b88d6d61c81d1000000"}`))
	}))
	defer server.Close()

	_, err := getManifest(server.URL)
	if err != nil {
		t.Errorf("getManifest() returned an error: %v", err)
	}
}

func TestDownload(t *testing.T) {
	expected := []byte(`Hello, World!`)
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		gw := gzip.NewWriter(rw)
		defer gw.Close()

		_, err := gw.Write(expected)
		if err != nil {
			http.Error(rw, "Failed to write response", http.StatusInternalServerError)
		}
	}))
	defer server.Close()

	m := &manifest{
		Version: "1.0.0",
		Sha256:  "2f4de17be8315ecf786e320d71c5a18e3096bf75b618ff4ead1b013fe04b761b",
		base:    server.URL,
	}

	actual, err := m.download()
	if err != nil {
		t.Errorf("download() returned an error: %v", err)
	}
	if !bytes.Equal(expected, actual) {
		t.Errorf("download() returned %v, expected %v", actual, expected)
	}
}
