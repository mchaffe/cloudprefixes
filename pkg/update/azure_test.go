package update

import (
	"fmt"
	"os"
	"testing"

	"github.com/mchaffe/cloudprefixes/pkg/db"
)

func TestMicrosoftURLFinder_FindURL(t *testing.T) {
	html, err := os.ReadFile("testdata/azure_response.html")
	if err != nil {
		t.Error(err)
	}
	tests := []struct {
		name    string
		f       *MicrosoftURLFinder
		body    []byte
		want    string
		wantErr bool
	}{
		{
			"Match found",
			&MicrosoftURLFinder{},
			html,
			"https://download.microsoft.com/download/7/1/D/71D86715-5596-4529-9B13-DA13A5DE5B63/ServiceTags_Public_20240923.json",
			false,
		},
		{
			"No Match",
			&MicrosoftURLFinder{},
			[]byte("<html>nope</html>"),
			"",
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.f.FindURL(tt.body)
			if (err != nil) != tt.wantErr {
				t.Errorf("MicrosoftURLFinder.FindURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("MicrosoftURLFinder.FindURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUpdateManager_UpdateAzurePrefixes(t *testing.T) {
	// Create a new test server
	ts := NewTestServer()
	defer ts.Close()

	// Mock function for GetJsonUrl
	mockGetJsonUrl := func(url string, finder URLFinder) (string, error) {
		return fmt.Sprintf("%s/azure_response.json", ts.URL()), nil
	}

	// Create a temporary file for the database
	f, err := os.CreateTemp("", "test_cloudprefixes.db")
	if err != nil {
		panic(err)
	}
	defer os.Remove((f.Name()))

	// Initialize IPRangeManager with the temporary database file
	dm, err := db.NewPrefixManager(f.Name())
	if err != nil {
		panic(err)
	}

	// Create the UpdateManager
	manager := &UpdateManager{
		PrefixManager: dm,
		GetJsonUrl:    mockGetJsonUrl,
	}

	tests := []struct {
		name    string
		m       *UpdateManager
		url     string
		wantErr bool
	}{
		{"working", manager, ts.URL() + "/azure_response.html", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.m.UpdateAzurePrefixes(tt.url); (err != nil) != tt.wantErr {
				t.Errorf("UpdateManager.UpdateAzurePrefixes() error = %v, wantErr %v", err, tt.wantErr)
			}
			found, prefixes, err := manager.PrefixManager.ContainsIP("13.69.109.133")
			if err != nil {
				t.Fatalf("failed to query prefixes: %v", err)
			}

			if !found && len(prefixes) != 4 {
				t.Errorf("UpdateManager.UpdateAzurePrefixes() len = %d, wanted 4", len(prefixes))
			}
		})
	}
}
