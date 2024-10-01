package update

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"regexp"

	"github.com/mchaffe/cloudprefixes/pkg/db"
)

type AzureResponse struct {
	ChangeNumber int    `json:"changeNumber"`
	Cloud        string `json:"cloud"`
	Values       []struct {
		Name            string   `json:"name"`
		ID              string   `json:"id"`
		NetworkFeatures []string `json:"networkFeatures"`
		Properties      struct {
			ChangeNumber    int      `json:"changeNumber"`
			Region          *string  `json:"region"`
			RegionID        int      `json:"regionId"`
			Platform        string   `json:"platform"`
			SystemService   *string  `json:"systemService"`
			AddressPrefixes []string `json:"addressPrefixes"`
		} `json:"properties"`
	} `json:"values"`
}

type URLFinder interface {
	FindURL(body []byte) (string, error)
}

type MicrosoftURLFinder struct{}

func (f *MicrosoftURLFinder) FindURL(body []byte) (string, error) {
	re := regexp.MustCompile(`(?m)https:\/\/download\.microsoft\.com\/download\/.*?\.json`)
	match := re.Find(body)
	if match == nil {
		slog.Debug("HTML body", "body", string(body))
		return "", fmt.Errorf("unable to match URL in html response")
	}
	return string(match), nil
}

func (m *UpdateManager) UpdateAzurePrefixes(url string) error {
	slog.Info("fetching HTML to find JSON", "url", url)
	jsonUrl, err := GetJsonUrl(url, &MicrosoftURLFinder{})
	if err != nil {
		return err
	}

	slog.Info("fetching JSON", "url", url)
	body, err := GetJson(jsonUrl)
	if err != nil {
		return err
	}

	var j AzureResponse
	err = json.Unmarshal(body, &j)
	if err != nil {
		return err
	}

	var prefixes []db.PrefixInfo
	for _, value := range j.Values {
		region := value.Properties.Region
		if region != nil {
			*region = "global"
		}
		for _, addressPrefix := range value.Properties.AddressPrefixes {
			prefixes = append(prefixes, db.PrefixInfo{
				Platform: "Azure",
				Region:   region,
				Service:  value.Properties.SystemService,
				Prefix:   addressPrefix,
			})
		}
	}
	return m.InsertPrefixes(prefixes)
}
