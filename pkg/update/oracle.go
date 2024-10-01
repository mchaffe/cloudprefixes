package update

import (
	"encoding/json"

	"github.com/mchaffe/cloudprefixes/pkg/db"
)

type OracleResponse struct {
	LastUpdatedTimestamp string `json:"last_updated_timestamp"`
	Regions              []struct {
		Region *string `json:"region"`
		Cidrs  []struct {
			Cidr string   `json:"cidr"`
			Tags []string `json:"tags"`
		} `json:"cidrs"`
	} `json:"regions"`
}

func (m *UpdateManager) UpdateOraclePrefixes(url string) error {
	body, err := GetJson(url)
	if err != nil {
		return err
	}

	var j OracleResponse
	err = json.Unmarshal(body, &j)
	if err != nil {
		return err
	}

	var prefixes []db.PrefixInfo
	for _, r := range j.Regions {
		for _, c := range r.Cidrs {
			for _, t := range c.Tags {
				prefixes = append(prefixes, db.PrefixInfo{
					Platform: "Oracle",
					Region:   r.Region,
					Prefix:   c.Cidr,
					Service:  &t,
				})
			}
		}
	}

	return m.InsertPrefixes(prefixes)
}
