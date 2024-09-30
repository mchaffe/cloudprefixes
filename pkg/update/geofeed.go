package update

import (
	"cloudprefixes/pkg/db"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Geofeed struct {
	Prefix      string  `json:"-"`
	CountryCode *string `json:"country_code,omitempty"`
	RegionCode  *string `json:"region_code,omitempty"`
	City        *string `json:"city,omitempty"`
	Postal      *string `json:"postal,omitempty"`
}

// Helper function to handle optional string fields in the CSV
func optionalString(record []string, index int) *string {
	if len(record) > index && record[index] != "" {
		return &record[index]
	}
	return nil
}

func (m *UpdateManager) UpdateGeoFeedPrefixes(url string, platform string) error {
	res, err := http.Get(url)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
	}

	var prefixes []db.PrefixInfo

	reader := csv.NewReader(res.Body)
	reader.Comma = ','

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("error reading CSV: %v", err)
		}

		if len(record) == 0 {
			continue
		}

		location := Geofeed{
			CountryCode: optionalString(record, 1),
			RegionCode:  optionalString(record, 2),
			City:        optionalString(record, 3),
			Postal:      optionalString(record, 4),
		}

		metaMap := map[string]Geofeed{
			"location": location,
		}
		metaJSON, err := json.Marshal(metaMap)
		if err != nil {
			return err
		}
		metaStr := string(metaJSON)

		prefixes = append(prefixes, db.PrefixInfo{
			Prefix:   record[0],
			Platform: platform,
			Metadata: &metaStr,
		})
	}

	return m.InsertPrefixes(prefixes)
}
