package webhook

import (
	"context"
	"fmt"
	"strings"

	"github.com/jetstack/cert-manager/pkg/acme/webhook"
	acme "github.com/jetstack/cert-manager/pkg/acme/webhook/apis/acme/v1alpha1"
	"github.com/jetstack/cert-manager/pkg/issuer/acme/dns/util"
	"github.com/nrdcg/dnspod-go"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog"
)

// Solver creates a new solver
func Solver() webhook.Solver {
	return &solver{}
}

// solver implements the provider-specific logic needed to
// 'present' an ACME challenge TXT record for your own DNS provider.
// To do so, it must implement the `github.com/jetstack/cert-manager/pkg/acme/webhook.Solver`
// interface.
type solver struct {
	client *kubernetes.Clientset

	dnspod map[int]*dnspod.Client
}

// Name is used as the name for this DNS solver when referencing it on the ACME
// Issuer resource.
// This should be unique **within the group name**, i.e. you can have two
// solvers configured with the same Name() **so long as they do not co-exist
// within a single webhook deployment**.
func (c *solver) Name() string {
	return "dnspod"
}

// Initialize will be called when the webhook first starts.
// This method can be used to instantiate the webhook, i.e. initialising
// connections or warming up caches.
// Typically, the kubeClientConfig parameter is used to build a Kubernetes
// client that can be used to fetch resources from the Kubernetes API, e.g.
// Secret resources containing credentials used to authenticate with DNS
// provider accounts.
// The stopCh can be used to handle early termination of the webhook, in cases
// where a SIGTERM or similar signal is sent to the webhook process.
func (c *solver) Initialize(kubeClientConfig *rest.Config, stopCh <-chan struct{}) error {
	cl, err := kubernetes.NewForConfig(kubeClientConfig)
	if err != nil {
		return err
	}
	c.client = cl

	c.dnspod = make(map[int]*dnspod.Client)

	return nil
}

// Present is responsible for actually presenting the DNS record with the
// DNS provider.
// This method should tolerate being called multiple times with the same value.
// cert-manager itself will later perform a self check to ensure that the
// solver has correctly configured the DNS provider.
func (c *solver) Present(ch *acme.ChallengeRequest) error {
	client, cfg, err := c.dnspodClient(ch)
	if err != nil {
		return err
	}

	domainID, err := getDomainID(client, ch.ResolvedZone)
	if err != nil {
		return err
	}

	recordAttributes := newTxtRecord(ch.ResolvedZone, ch.ResolvedFQDN, ch.Key, *cfg.TTL)
	_, _, err = client.Records.Create(domainID, *recordAttributes)
	if err != nil {
		return fmt.Errorf("dnspod API call failed: %v", err)
	}

	return nil
}

// CleanUp should delete the relevant TXT record from the DNS provider console.
// If multiple TXT records exist with the same record name (e.g.
// _acme-challenge.example.com) then **only** the record with the same `key`
// value provided on the ChallengeRequest should be cleaned up.
// This is in order to facilitate multiple DNS validations for the same domain
// concurrently.
func (c *solver) CleanUp(ch *acme.ChallengeRequest) error {

	client, _, err := c.dnspodClient(ch)
	if err != nil {
		return err
	}

	domainID, err := getDomainID(client, ch.ResolvedZone)
	if err != nil {
		return err
	}

	records, err := findTxtRecords(client, domainID, ch.ResolvedZone, ch.ResolvedFQDN)
	if err != nil && !strings.Contains(err.Error(), "No records") {
		return err
	}

	for _, record := range records {
		if record.Value != ch.Key {
			continue
		}

		_, err := client.Records.Delete(domainID, record.ID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *solver) dnspodClient(ch *acme.ChallengeRequest) (*dnspod.Client, config, error) {
	cfg, err := loadConfig(ch.Config)

	if err != nil {
		return nil, cfg, err
	}

	klog.Infof("config: %v", cfg)

	apiID := cfg.APIID

	client, ok := c.dnspod[apiID]

	if ok {
		return client, cfg, nil
	}

	ref := cfg.APITokenSecretRef

	secret, err := c.client.CoreV1().Secrets(ch.ResourceNamespace).Get(context.TODO(), ref.Name, meta.GetOptions{})

	if err != nil {
		return nil, cfg, err
	}

	apiToken, ok := secret.Data[ref.Key]
	if !ok {
		return nil, cfg, fmt.Errorf("no api token for %q in secret '%s/%s'", ref.Name, ref.Key, ch.ResourceNamespace)
	}

	key := fmt.Sprintf("%d,%s", cfg.APIID, apiToken)
	params := dnspod.CommonParams{LoginToken: key, Format: "json"}

	client = dnspod.NewClient(params)
	c.dnspod[apiID] = client

	return client, cfg, nil
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
