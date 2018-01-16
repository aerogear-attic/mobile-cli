package integration

import (
	"io/ioutil"
	"path/filepath"
	"runtime"
	"testing"
)

func snapshotPath(t *testing.T, snapshot string) string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatalf("problems recovering caller information")
	}

	return filepath.Join(filepath.Dir(filename), snapshot)
}

func WriteSnapshot(t *testing.T, snapshot string, content []byte) {
	snapshotPath := snapshotPath(t, snapshot)
	if err := ioutil.WriteFile(snapshotPath, content, 0644); err != nil {
		t.Fatalf("Unexpected error writing snapshot %v", err)
	}
}

func LoadSnapshot(t *testing.T, snapshot string) string {
	snapshotPath := snapshotPath(t, snapshot)
	content, err := ioutil.ReadFile(snapshotPath)
	if err != nil {
		t.Fatalf("Unexpected error loading snapshot %v", err)
	}

	return string(content)
}
