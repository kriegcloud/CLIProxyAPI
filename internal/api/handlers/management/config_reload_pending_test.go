package management

import (
	"context"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/config"
	"testing"
)

func TestSetConfigCannotOverwritePendingManagementSave(t *testing.T) {
	handler := NewHandler(&config.Config{Port: 8411}, "", nil)
	handler.mu.Lock()
	snapshot := handler.reloadSnapshotConfigLocked()
	handler.mu.Unlock()
	handler.SetConfig(&config.Config{Port: 8400})
	if handler.cfg.Port != 8411 {
		t.Fatalf("pending save overwritten: port %d", handler.cfg.Port)
	}
	handler.reloadConfigAfterManagementSave(context.Background(), snapshot)
	handler.SetConfig(&config.Config{Port: 8422})
	if handler.cfg.Port != 8422 {
		t.Fatalf("later config update blocked: port %d", handler.cfg.Port)
	}
}
