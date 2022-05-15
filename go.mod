module github.com/kaelzhang/cert-manager-webhook-dnspod

go 1.17

require (
	github.com/jetstack/cert-manager v1.7.0
	github.com/nrdcg/dnspod-go v0.4.1-0.20201003132448-1ad9f84ef877
	k8s.io/apiextensions-apiserver v0.19.4
	k8s.io/apimachinery v0.19.4
	k8s.io/client-go v0.23.1
	k8s.io/klog/v2 v2.4.0
)
