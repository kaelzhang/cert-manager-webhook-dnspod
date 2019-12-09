module github.com/kaelzhang/cert-manager-webhook-dnspod

go 1.13

require (
	github.com/decker502/dnspod-go v0.2.0
	github.com/jetstack/cert-manager v0.12.0

	k8s.io/apiextensions-apiserver v0.0.0-20191114105449-027877536833
	k8s.io/apimachinery v0.0.0-20191028221656-72ed19daf4bb
	k8s.io/client-go v0.0.0-20191114101535-6c5935290e33
)

// We need to hold back client_golang to prevent issue
// ```
// metrics/legacyregistry/registry.go:44:9: undefined: prometheus.InstrumentHandler
// ```
// Ref
// https://github.com/jetstack/cert-manager/issues/2432
replace github.com/prometheus/client_golang => github.com/prometheus/client_golang v0.9.4
