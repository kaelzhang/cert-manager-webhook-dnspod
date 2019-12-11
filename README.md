# Cert-Manager ACME webhook for DNSPod

> A fork of [qqshfox/cert-manager-webhook-dnspod](https://github.com/qqshfox/cert-manager-webhook-dnspod), and is updated to cert-manager >= 0.12.0

This is a webhook solver for [DNSPod](https://www.dnspod.cn).

## Prerequisites

- [cert-manager](https://github.com/jetstack/cert-manager): >= 0.12.0
  - [Installing on Kubernetes](https://cert-manager.io/docs/installation/kubernetes/)

## Installation

```console
$ helm install cert-manager-webhook-dnspod ./charts
```

### Prepare for DNSPod

- Generate API ID and API Token from DNSPod (https://support.dnspod.cn/Kb/showarticle/tsid/227/)

- Create secret to store the API Token

```sh
kubectl --namespace cert-manager create secret generic \
    dnspod-credentials --from-literal=api-token='<DNSPOD_API_TOKEN>'
```

### Issuer

Create a production issuer. And you could create a staging letsencrypt issuer if necessary.

```yaml
apiVersion: cert-manager.io/v1alpha2
kind: Issuer
metadata:
  name: letsencrypt-prod
spec:
  acme:
    # The ACME server URL
    server: https://acme-v02.api.letsencrypt.org/directory

    # Email address used for ACME registration
    email: <your email>

    # Name of a secret used to store the ACME account private key
    privateKeySecretRef:
      name: letsencrypt-prod

    solvers:
    - dns01:
        webhook:
          groupName: <your group>
          solverName: dnspod
          config:
            apiID: <your dnspod api id>
            apiTokenSecretRef:
              key: api-token
              name: dnspod-credentials
```

### Certificate

```yaml
apiVersion: cert-manager.io/v1alpha2
kind: Certificate
metadata:
  # you could replace this name to your own
  name: wildcard-yourdomain-com # for *.yourdomain.com
spec:
  secretName: wildcard-yourdomain-com-tls
  renewBefore: 240h
  dnsNames:
    - '*.yourdomain.com'
  issuerRef:
    name: letsencrypt-prod
    kind: ClusterIssuer
```

### Ingress

A common use-case for cert-manager is requesting TLS signed certificates to secure your ingress resources. This can be done by simply adding annotations to your Ingress resources and cert-manager will facilitate creating the Certificate resource for you. A small sub-component of cert-manager, ingress-shim, is responsible for this.

For details, see [here](https://cert-manager.io/docs/usage/ingress/)

```yaml
apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  name: demo-ingress
  namespace: default
  annotations:
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
spec:
  tls:
  - hosts:
    - '*.yourdomain.com'
    secretName: wildcard-yourdomain-com-tls
  rules:
  - host: demo.yourdomain.com
    http:
      paths:
      - path: /
        backend:
          serviceName: backend-service
          servicePort: 80
```

## Development

All DNS providers **must** run the DNS01 provider conformance testing suite,
else they will have undetermined behaviour when used with cert-manager.

**It is essential that you configure and run the test suite when creating a
DNS01 webhook.**

An example Go test file has been provided in [main_test.go]().

Before you can run the test suite, you need to download the test binaries:

```sh
mkdir __main__
wget -O- https://storage.googleapis.com/kubebuilder-tools/kubebuilder-tools-1.14.1-darwin-amd64.tar.gz | tar x -
mv kubebuilder __main__/hack
```

Then modify `testdata/my-custom-solver/config.json` to setup the configs.

Now we could run tests in debug mode with dlv

```sh
GROUP_NAME=ost.ai \
TEST_ZONE_NAME=ost.ai. \
dlv test . -- -test.v
```
