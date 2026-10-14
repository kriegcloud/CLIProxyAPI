package watcher

import (
	"github.com/router-for-me/CLIProxyAPI/v8/internal/config"
	"os"
	"path/filepath"
	"testing"
)

func TestReloadDoesNotAcknowledgeLaterUnappliedDiskRevision(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, "config.yaml")
	if err := os.WriteFile(path, []byte("port: 8411\n"), 0600); err != nil {
		t.Fatal(err)
	}
	var seen []int
	watcher, err := NewWatcher(path, directory, func(cfg *config.Config) {
		seen = append(seen, cfg.Port)
		if len(seen) == 1 {
			if err := os.WriteFile(path, []byte("port: 8422\n"), 0600); err != nil {
				t.Fatal(err)
			}
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if errClose := watcher.watcher.Close(); errClose != nil {
			t.Error(errClose)
		}
	}()
	watcher.reloadConfigIfChanged()
	watcher.reloadConfigIfChanged()
	if len(seen) != 2 || seen[0] != 8411 || seen[1] != 8422 {
		t.Fatalf("applied revisions = %v", seen)
	}
}
