package dnspod

import (
	"context"
	"fmt"
	"strings"

	acme "github.com/jetstack/cert-manager/pkg/acme/webhook/apis/acme/v1alpha1"
	"github.com/nrdcg/dnspod-go"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	klog "k8s.io/klog/v2"
)

// Solver implements the provider-specific logic needed to
// 'present' an ACME challenge TXT record for your own DNS provider.
// To do so, it must implement the `github.com/jetstack/cert-manager/pkg/acme/webhook.Solver`
// interface.
type Solver struct {
	client *kubernetes.Clientset

	dnspod map[int]*dnspod.Client
}

// Name is used as the name for this DNS solver when referencing it on the ACME
// Issuer resource.
// This should be unique **within the group name**, i.e. you can have two
// solvers configured with the same Name() **so long as they do not co-exist
// within a single webhook deployment**.
func (c *Solver) Name() string {
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
func (c *Solver) Initialize(kubeClientConfig *rest.Config, stopCh <-chan struct{}) error {
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
func (c *Solver) Present(ch *acme.ChallengeRequest) error {
	// return fmt.Errorf("just for debug Present: ch: %v", ch)

	client, cfg, err := c.dnspodClient(ch)
	if err != nil {
		klog.Errorf("Present: fails to initialize dnspod client from challenge: %v", err)

		return err
	}

	domainID, err := getDomainID(client, ch.ResolvedZone)
	if err != nil {

		klog.Errorf("Present: fails to get domain id for resolved zone (%v): %v", ch.ResolvedZone, err)

		return err
	}

	recordAttributes := newTxtRecord(ch.ResolvedZone, ch.ResolvedFQDN, ch.Key, *cfg.TTL)

	klog.Errorf("Present: recordAttributes: %v, %v", recordAttributes, ch.Key)

	_, _, err = client.Records.Create(domainID, *recordAttributes)
	if err != nil {
		klog.Errorf("Present: fails to add txt record: %v", err)

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
func (c *Solver) CleanUp(ch *acme.ChallengeRequest) error {
	// return fmt.Errorf("just for debug CleanUp: ch: %v", ch)

	client, _, err := c.dnspodClient(ch)
	if err != nil {
		klog.Errorf("CleanUp: fails to initialize dnspod client from challenge: %v", err)

		return err
	}

	domainID, err := getDomainID(client, ch.ResolvedZone)
	if err != nil {
		klog.Errorf("CleanUp: fails to get domain id for resolved zone `%v`: %v", ch.ResolvedZone, err)

		return err
	}

	records, err := findTxtRecords(client, domainID, ch.ResolvedZone, ch.ResolvedFQDN)
	if err != nil && !strings.Contains(err.Error(), "No records") {
		klog.Errorf("CleanUp: fails to find txt record (%v, %v, %v): %v", domainID, ch.ResolvedZone, ch.ResolvedFQDN, err)

		return err
	}

	for _, record := range records {
		if record.Value != ch.Key {
			continue
		}

		_, err := client.Records.Delete(domainID, record.ID)
		if err != nil {
			klog.Errorf("CleanUp: fails to delete txt record (%v, %v): %v", domainID, record.ID, err)

			return err
		}
	}

	return nil
}

func (c *Solver) dnspodClient(ch *acme.ChallengeRequest) (*dnspod.Client, config, error) {
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
