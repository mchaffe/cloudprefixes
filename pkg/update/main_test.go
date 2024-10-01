package update

import (
	"database/sql"
	"fmt"
	"strings"
	"testing"

	"github.com/mchaffe/cloudprefixes/pkg/db"
)

type MockMicrosoftURLFinder struct{}

func (m *MockMicrosoftURLFinder) GetJsonUrl(url string) (string, error) {
	if url == "invalid" {
		return "", fmt.Errorf("unable to match URL in html response")
	}
	return strings.TrimSuffix(url, "html") + "json", nil
}

func TestUpdateManager_InsertPrefixes(t *testing.T) {
	// Initialize an in-memory SQLite database for testing
	dbConn, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open in-memory SQLite database: %v", err)
	}
	defer dbConn.Close()

	// Create a PrefixManager instance with the test database
	prefixManager, err := db.NewIPRangeManager(":memory:") // Use in-memory DB for tests
	if err != nil {
		t.Fatalf("Failed to initialize PrefixManager: %v", err)
	}
	defer prefixManager.Close()

	// Initialize UpdateManager with the PrefixManager
	updateManager := NewUpdateManager(prefixManager)

	tests := []struct {
		name     string
		prefixes []db.PrefixInfo
		wantErr  bool
	}{
		{
			name: "Insert valid prefixes",
			prefixes: []db.PrefixInfo{
				{
					Prefix:   "192.168.1.0/24",
					Platform: "AWS",
					Region:   stringPointer("us-east-1"),
				},
				{
					Prefix:   "2001:db8::/32",
					Platform: "GCP",
					Region:   stringPointer("global"),
				},
			},
			wantErr: false,
		},
		{
			name:     "Insert empty prefixes",
			prefixes: []db.PrefixInfo{},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := updateManager.InsertPrefixes(tt.prefixes)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateManager.InsertPrefixes() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
