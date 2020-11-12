module github.com/kaelzhang/cert-manager-webhook-dnspod

go 1.15

require (
	github.com/jetstack/cert-manager v1.0.4
	github.com/nrdcg/dnspod-go v0.4.1-0.20201003132448-1ad9f84ef877
	k8s.io/apiextensions-apiserver v0.19.3
	k8s.io/apimachinery v0.19.3
	k8s.io/client-go v0.19.3
)

// We need to hold back client_golang to prevent issue
// ```
// metrics/legacyregistry/registry.go:44:9: undefined: prometheus.InstrumentHandler
// ```
// Ref
// https://github.com/jetstack/cert-manager/issues/2432
// replace github.com/prometheus/client_golang => github.com/prometheus/client_golang v0.9.4
