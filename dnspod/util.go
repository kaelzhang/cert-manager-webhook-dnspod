package dnspod

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jetstack/cert-manager/pkg/issuer/acme/dns/util"
	"github.com/nrdcg/dnspod-go"
	extapi "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
)

// loadConfig is a small helper function that decodes JSON configuration into
// the typed config struct.
func loadConfig(cfgJSON *extapi.JSON) (config, error) {
	ttl := defaultTTL
	cfg := config{TTL: &ttl}
	// handle the 'base case' where no configuration has been provided
	if cfgJSON == nil {
		return cfg, nil
	}
	if err := json.Unmarshal(cfgJSON.Raw, &cfg); err != nil {
		return cfg, fmt.Errorf("error decoding solver config: %v", err)
	}

	return cfg, nil
}

func getDomainID(client *dnspod.Client, zone string) (string, error) {
	domains, _, err := client.Domains.List()
	if err != nil {
		return "", fmt.Errorf("dnspod API call failed: %v", err)
	}

	authZone, err := util.FindZoneByFqdn(zone, util.RecursiveNameservers)
	if err != nil {
		return "", err
	}

	var hostedDomain dnspod.Domain
	for _, domain := range domains {
		if domain.Name == util.UnFqdn(authZone) {
			hostedDomain = domain
			break
		}
	}

	hostedDomainID, err := hostedDomain.ID.Int64()
	if err != nil {
		return "", err
	}
	if hostedDomainID == 0 {
		return "", fmt.Errorf("Zone %s not found in dnspod for zone %s", authZone, zone)
	}

	return fmt.Sprintf("%d", hostedDomainID), nil
}

func newTxtRecord(zone, fqdn, value string, ttl int) *dnspod.Record {
	name := extractRecordName(fqdn, zone)

	return &dnspod.Record{
		Type:  "TXT",
		Name:  name,
		Value: value,
		Line:  "默认",
		TTL:   fmt.Sprintf("%d", ttl),
	}
}

func findTxtRecords(client *dnspod.Client, domainID, zone, fqdn string) ([]dnspod.Record, error) {
	recordName := extractRecordName(fqdn, zone)
	records, _, err := client.Records.List(domainID, recordName)
	if err != nil {
		return records, fmt.Errorf("dnspod API call has failed: %v", err)
	}

	return records, nil
}

func extractRecordName(fqdn, zone string) string {
	if idx := strings.Index(fqdn, "."+zone); idx != -1 {
		return fqdn[:idx]
	}

	return util.UnFqdn(fqdn)
}
