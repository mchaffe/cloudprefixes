package update

import (
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"

	"github.com/mchaffe/cloudprefixes/pkg/db"
)

type UpdateManager struct {
	PrefixManager *db.PrefixManager
	GetJsonUrl    func(string, URLFinder) (string, error) // Dependency injection
}

func NewUpdateManager(prefixManager *db.PrefixManager) *UpdateManager {
	return &UpdateManager{PrefixManager: prefixManager}
}

func GetJsonUrl(url string, finder URLFinder) (string, error) {
	res, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return "", fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("error reading JSON response: %v", err)
	}

	return finder.FindURL(body)
}

func GetJson(url string) (body []byte, err error) {
	res, err := http.Get(url)
	if err != nil {
		return []byte{}, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return []byte{}, fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
	}

	body, err = io.ReadAll(res.Body)
	if err != nil {
		return []byte{}, fmt.Errorf("error reading response: %v", err)
	}
	return body, nil
}

func (m *UpdateManager) InsertPrefixes(prefixes []db.PrefixInfo) error {
	err := m.PrefixManager.AddPrefixBatch(prefixes)
	if err != nil {
		return fmt.Errorf("error inserting data %v", err)
	}
	slog.Info("successfully inserted prefixes", "count", len(prefixes))
	return nil
}

func (m *UpdateManager) UpdateAllSources() {
	err := m.PrefixManager.ClearAllData()
	if err != nil {
		log.Fatalf("failed to clear existing data: %v", err)
	}

	slog.Info("Updating prefixes: GitHub")
	err = m.UpdateGithubPrefixes("https://api.github.com/meta")
	if err != nil {
		log.Fatal(err)
	}

	slog.Info("Updating prefixes: Azure public")
	err = m.UpdateAzurePrefixes("https://www.microsoft.com/en-us/download/details.aspx?id=56519")
	if err != nil {
		log.Fatal(err)
	}
	slog.Info("Updating prefixes: Azure US government")
	err = m.UpdateAzurePrefixes("https://www.microsoft.com/en-us/download/details.aspx?id=57063")
	if err != nil {
		log.Fatal(err)
	}
	slog.Info("Updating prefixes: Azure China")
	err = m.UpdateAzurePrefixes("https://www.microsoft.com/en-us/download/details.aspx?id=57062")
	if err != nil {
		log.Fatal(err)
	}
	slog.Info("Updating prefixes: Azure Germany")
	err = m.UpdateAzurePrefixes("https://www.microsoft.com/en-au/download/details.aspx?id=57064")
	if err != nil {
		log.Fatal(err)
	}

	slog.Info("Updating prefixes: AWS")
	err = m.UpdateAwsPrefixes("https://ip-ranges.amazonaws.com/ip-ranges.json")
	if err != nil {
		log.Fatal(err)
	}

	slog.Info("Updating prefixes: GCP")
	err = m.UpdateGooglePrefixes("https://www.gstatic.com/ipranges/cloud.json", "GCP")
	if err != nil {
		log.Fatal(err)
	}

	slog.Info("Updating prefixes: Google")
	err = m.UpdateGooglePrefixes("https://www.gstatic.com/ipranges/goog.json", "Google")
	if err != nil {
		log.Fatal(err)
	}

	slog.Info("Updating prefixes: Oracle")
	err = m.UpdateOraclePrefixes("https://docs.oracle.com/en-us/iaas/tools/public_ip_ranges.json")
	if err != nil {
		log.Fatal(err)
	}

	geofeeds := []struct {
		url  string
		name string
	}{
		{name: "Digial Ocean", url: "https://digitalocean.com/geo/google.csv"},
		{name: "CloudFlare", url: "https://www.cloudflare.com/ips-v4"},
		{name: "CloudFlare", url: "https://www.cloudflare.com/ips-v6"},
	}
	for _, g := range geofeeds {
		slog.Info("Updating prefixes:", "geofeed", g.name)
		err = m.UpdateGeoFeedPrefixes(g.url, g.name)
		if err != nil {
			log.Fatal(err)
		}
	}
}
