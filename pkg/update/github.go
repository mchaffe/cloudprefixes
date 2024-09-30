package update

import (
	"cloudprefixes/pkg/db"
	"encoding/json"
	"net"
	"reflect"

	"log/slog"
)

type GithubResponse struct {
	VerifiablePasswordAuthentication bool `json:"verifiable_password_authentication"`
	SSHKeyFingerprints               struct {
		Sha256Ecdsa   string `json:"SHA256_ECDSA"`
		Sha256Ed25519 string `json:"SHA256_ED25519"`
		Sha256Rsa     string `json:"SHA256_RSA"`
	} `json:"ssh_key_fingerprints"`
	SSHKeys                  []string `json:"ssh_keys"`
	Hooks                    []string `json:"hooks"`
	Web                      []string `json:"web"`
	API                      []string `json:"api"`
	Git                      []string `json:"git"`
	GithubEnterpriseImporter []string `json:"github_enterprise_importer"`
	Packages                 []string `json:"packages"`
	Pages                    []string `json:"pages"`
	Importer                 []string `json:"importer"`
	Actions                  []string `json:"actions"`
	ActionsMacos             []string `json:"actions_macos"`
	Codespaces               []string `json:"codespaces"`
	Dependabot               []string `json:"dependabot"`
	Copilot                  []string `json:"copilot"`
	Domains                  struct {
		Website              []string `json:"website"`
		Codespaces           []string `json:"codespaces"`
		Copilot              []string `json:"copilot"`
		Packages             []string `json:"packages"`
		Actions              []string `json:"actions"`
		ArtifactAttestations struct {
			TrustDomain string   `json:"trust_domain"`
			Services    []string `json:"services"`
		} `json:"artifact_attestations"`
	} `json:"domains"`
}

func iterateCIDRFields(g GithubResponse) (prefixes []db.PrefixInfo) {
	v := reflect.ValueOf(g)
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		if field.Kind() == reflect.Slice && field.Type().Elem().Kind() == reflect.String {
			slice := field.Interface().([]string)
			if len(slice) > 0 && isCIDR(slice[0]) {
				slog.Info("Field contains CIDRs:", "field", fieldType.Name)
				for _, cidr := range slice {
					prefixes = append(prefixes, db.PrefixInfo{
						Platform: "GitHub",
						Prefix:   cidr,
						Service:  &fieldType.Name,
					})
				}
			}
		}
	}
	return prefixes
}

func isCIDR(s string) bool {
	_, _, err := net.ParseCIDR(s)
	return err == nil
}

func (m *UpdateManager) UpdateGithubPrefixes(url string) error {
	body, err := GetJson(url)
	if err != nil {
		return err
	}

	var j GithubResponse
	err = json.Unmarshal(body, &j)
	if err != nil {
		return err
	}

	prefixes := iterateCIDRFields(j)

	return m.InsertPrefixes(prefixes)
}
