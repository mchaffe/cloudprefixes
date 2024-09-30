package update

import (
	"testing"
)

func TestUpdateManager_UpdateGithubPrefixes(t *testing.T) {
	manager, ts, cleanup := SetupUpdateManager()
	defer cleanup()

	tests := []struct {
		name    string
		m       *UpdateManager
		url     string
		wantErr bool
	}{
		{"working", manager, ts.URL() + "/github_response.json", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.m.UpdateGithubPrefixes(tt.url); (err != nil) != tt.wantErr {
				t.Errorf("UpdateManager.UpdateGithubPrefixes() error = %v, wantErr %v", err, tt.wantErr)
			}
			found, prefixes, err := manager.PrefixManager.ContainsIP("13.69.109.133")
			if err != nil {
				t.Fatalf("failed to query prefixes: %v", err)
			}

			if !found && len(prefixes) != 1 {
				t.Errorf("UpdateManager.UpdateAzurePrefixes() len = %d, wanted 1", len(prefixes))
			}
		})
	}
}
