package update

import (
	"cloudprefixes/pkg/db"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
)

func stringPointer(s string) *string {
	return &s
}

// TestServer represents a local HTTP server for testing
type TestServer struct {
	Server *httptest.Server
}

// NewTestServer creates and starts a new test server
func NewTestServer() *TestServer {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handleTestData)

	server := httptest.NewServer(mux)
	return &TestServer{Server: server}
}

// Close shuts down the test server
func (ts *TestServer) Close() {
	ts.Server.Close()
}

// URL returns the base URL of the test server
func (ts *TestServer) URL() string {
	return ts.Server.URL
}

// handleTestData serves files from the testdata directory
func handleTestData(w http.ResponseWriter, r *http.Request) {
	filePath := filepath.Join("testdata", r.URL.Path)
	content, err := os.ReadFile(filePath)
	if err != nil {
		http.Error(w, fmt.Sprintf("File not found: %s", r.URL.Path), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(content)
}

func SetupUpdateManager() (*UpdateManager, *TestServer, func()) {
	// Create a new test server
	ts := NewTestServer()

	// Initialize IPRangeManager with the temporary database
	dm, err := db.NewIPRangeManager(":memory:")
	if err != nil {
		panic(err)
	}

	// Create the UpdateManager
	manager := &UpdateManager{
		PrefixManager: dm,
		GetJsonUrl:    GetJsonUrl,
	}

	// Return manager, test server, and cleanup function
	return manager, ts, func() {
		ts.Close()
	}
}
