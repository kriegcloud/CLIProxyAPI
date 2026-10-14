package watcher

import (
	"github.com/router-for-me/CLIProxyAPI/v8/internal/config"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/watcher/synthesizer"
	coreauth "github.com/router-for-me/CLIProxyAPI/v8/sdk/cliproxy/auth"
	"path/filepath"
	"testing"
)

func TestConfigRescanRetainsPersistedAuthUntilNewFileObservation(t *testing.T) {
	path := filepath.Join(t.TempDir(), "credential.json")
	current := &coreauth.Auth{ID: "credential", Prefix: "new", Attributes: map[string]string{coreauth.AttributePath: path}}
	watcher := &Watcher{config: &config.Config{}, currentAuths: map[string]*coreauth.Auth{current.ID: current}, fileObservations: map[string]uint64{path: 4}, persistedAuthObservations: map[string]uint64{current.ID: 4}}
	previous := snapshotCoreAuthsFunc
	defer func() { snapshotCoreAuthsFunc = previous }()
	snapshotCoreAuthsFunc = func(*config.Config, string, synthesizer.PluginAuthParser) []*coreauth.Auth {
		old := current.Clone()
		old.Prefix = "old"
		return []*coreauth.Auth{old}
	}
	watcher.refreshAuthState(false)
	if watcher.currentAuths[current.ID].Prefix != "new" {
		t.Fatal("config rescan replaced persisted state")
	}
	watcher.observeAuthFile(path)
	watcher.refreshAuthState(false)
	if watcher.currentAuths[current.ID].Prefix != "old" {
		t.Fatal("new filesystem observation was ignored")
	}
}
