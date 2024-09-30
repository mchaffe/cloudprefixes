package update

import (
	"cloudprefixes/pkg/db"
	"encoding/json"
	"fmt"
)

type GoogleResponse struct {
	SyncToken    string `json:"synctoken"`
	CreationTime string `json:"creationTime"`
	Prefixes     []struct {
		Region     *string `json:"scaope"`
		Service    *string `json:"service"`
		IPv4Prefix string  `json:"ipv4Prefix"`
		IPv6Prefix string  `json:"ipv6Prefix"`
	} `json:"prefixes"`
}

func (m *UpdateManager) UpdateGooglePrefixes(url string, platform string) error {
	body, err := GetJson(url)
	if err != nil {
		return err
	}

	var j GoogleResponse
	err = json.Unmarshal(body, &j)
	if err != nil {
		return err
	}

	var prefixes []db.PrefixInfo
	for _, p := range j.Prefixes {
		var prefix string
		if p.IPv4Prefix != "" {
			prefix = p.IPv4Prefix
		} else if p.IPv6Prefix != "" {
			prefix = p.IPv6Prefix
		} else {
			return fmt.Errorf("unable to find prefix")
		}
		prefixes = append(prefixes, db.PrefixInfo{
			Platform: platform,
			Region:   p.Region,
			Service:  p.Service,
			Prefix:   prefix,
		})
	}

	return m.InsertPrefixes(prefixes)
}
