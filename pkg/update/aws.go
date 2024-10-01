package update

import (
	"encoding/json"

	"github.com/mchaffe/cloudprefixes/pkg/db"
)

type AwsResponse struct {
	SyncToken  string `json:"syncToken"`
	CreateDate string `json:"createDate"`
	Prefixes   []struct {
		IPPrefix           string  `json:"ip_prefix"`
		Region             *string `json:"region"`
		Service            *string `json:"service"`
		NetworkBorderGroup string  `json:"network_border_group"`
	} `json:"prefixes"`
	Ipv6Prefixes []struct {
		Ipv6Prefix         string  `json:"ipv6_prefix"`
		Region             *string `json:"region"`
		Service            *string `json:"service"`
		NetworkBorderGroup string  `json:"network_border_group"`
	} `json:"ipv6_prefixes"`
}

func (m *UpdateManager) UpdateAwsPrefixes(url string) error {
	body, err := GetJson(url)
	if err != nil {
		return err
	}

	var j AwsResponse
	err = json.Unmarshal(body, &j)
	if err != nil {
		return err
	}

	var prefixes []db.PrefixInfo
	for _, prefix := range j.Prefixes {
		metaMap := map[string]string{
			"network_boarder_group": prefix.NetworkBorderGroup,
		}
		metaJSON, err := json.Marshal(metaMap)
		if err != nil {
			return err
		}
		metaStr := string(metaJSON)

		prefixes = append(prefixes, db.PrefixInfo{
			Platform: "AWS",
			Region:   prefix.Region,
			Service:  prefix.Service,
			Prefix:   prefix.IPPrefix,
			Metadata: &metaStr,
		})
	}
	for _, prefix := range j.Ipv6Prefixes {
		metaMap := map[string]string{
			"network_boarder_group": prefix.NetworkBorderGroup,
		}
		metaJSON, err := json.Marshal(metaMap)
		if err != nil {
			return err
		}
		metaStr := string(metaJSON)

		prefixes = append(prefixes, db.PrefixInfo{
			Platform: "AWS",
			Region:   prefix.Region,
			Service:  prefix.Service,
			Prefix:   prefix.Ipv6Prefix,
			Metadata: &metaStr,
		})
	}

	return m.InsertPrefixes(prefixes)
}
