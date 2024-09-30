package update

import "testing"

func TestUpdateManager_UpdateOraclePrefixes(t *testing.T) {
	manager, ts, cleanup := SetupUpdateManager()
	defer cleanup()

	tests := []struct {
		name    string
		m       *UpdateManager
		url     string
		wantErr bool
	}{
		{"working", manager, ts.URL() + "/oracle_response.json", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.m.UpdateOraclePrefixes(tt.url); (err != nil) != tt.wantErr {
				t.Errorf("UpdateManager.UpdateOraclePrefixes() error = %v, wantErr %v", err, tt.wantErr)
			}
			found, prefixes, err := manager.PrefixManager.ContainsIP("134.70.8.1")
			if err != nil {
				t.Fatalf("failed to query prefixes: %v", err)
			}

			if !found && len(prefixes) != 2 {
				t.Errorf("UpdateManager.UpdateAzurePrefixes() len = %d, wanted 2", len(prefixes))
			}
		})
	}
}
