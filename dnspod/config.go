package dnspod

import (
	api "github.com/jetstack/cert-manager/pkg/apis/meta/v1"
)

const (
	defaultTTL = 600
)

// config is a structure that is used to decode into when
// solving a DNS01 challenge.
// This information is provided by cert-manager, and may be a reference to
// additional configuration that's needed to solve the challenge for this
// particular certificate or issuer.
// This typically includes references to Secret resources containing DNS
// provider credentials, in cases where a 'multi-tenant' DNS solver is being
// created.
// If you do *not* require per-issuer or per-certificate configuration to be
// provided to your webhook, you can skip decoding altogether in favour of
// using CLI flags or similar to provide configuration.
// You should not include sensitive information here. If credentials need to
// be used by your provider here, you should reference a Kubernetes Secret
// resource and fetch these credentials using a Kubernetes clientset.
type config struct {
	APIID             int                   `json:"apiID"`
	APITokenSecretRef api.SecretKeySelector `json:"apiTokenSecretRef"`
	TTL               *int                  `json:"ttl"`
}
