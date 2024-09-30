package update

import (
	"reflect"
	"testing"
)

func Test_optionalString(t *testing.T) {
	type args struct {
		record []string
		index  int
	}
	tests := []struct {
		name string
		args args
		want *string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := optionalString(tt.args.record, tt.args.index); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("optionalString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUpdateManager_UpdateGeoFeedPrefixes(t *testing.T) {
	manager, ts, cleanup := SetupUpdateManager()
	defer cleanup()

	tests := []struct {
		name    string
		m       *UpdateManager
		url     string
		wantErr bool
	}{
		{"working", manager, ts.URL() + "/digitalocean_response.csv", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.m.UpdateGeoFeedPrefixes(tt.url, "Digital Ocean"); (err != nil) != tt.wantErr {
				t.Errorf("UpdateManager.UpdateGeoFeedPrefixes() error = %v, wantErr %v", err, tt.wantErr)
			}
			found, prefixes, err := manager.PrefixManager.ContainsIP("45.55.32.1")
			if err != nil {
				t.Fatalf("failed to query prefixes: %v", err)
			}

			if !found && len(prefixes) != 1 {
				t.Errorf("UpdateManager.UpdateAzurePrefixes() len = %d, wanted 1", len(prefixes))
			}
		})
	}
}
