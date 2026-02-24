//go:build integration

package testinfra

import (
	"os"
	"path/filepath"
	"testing"
)

func ProjectRoot(t testing.TB) string {
	t.Helper()

	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("testinfra: getwd: %v", err)
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("testinfra: could not find project root (go.mod)")
		}
		dir = parent
	}
}
