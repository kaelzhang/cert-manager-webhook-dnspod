# Cert-Manager ACME webhook for DNSPod

> A fork of [qqshfox/cert-manager-webhook-dnspod](https://github.com/qqshfox/cert-manager-webhook-dnspod)

This is a webhook solver for Tencent [DNSPod](https://www.dnspod.cn).

Features
- Updated to cert-manager 1.0.4
- Updated to client-go 0.19.4

Tested on production environment of
- Kubernetes 1.18.3

## Prerequisites

Have [cert-manager](https://github.com/jetstack/cert-manager): >= 1.0.4 [installed](https://cert-manager.io/docs/installation/kubernetes/) within your kubernetes cluster.

## Installation

### Prepare for DNSPod

- Generate API ID and API Token from DNSPod (https://support.dnspod.cn/Kb/showarticle/tsid/227/)

- Create secret to store the API Token

```sh
kubectl --namespace cert-manager create secret generic \
  dnspod-credentials --from-literal=api-token='<DNSPOD_API_TOKEN>'
```

### Install `cert-manager-webhook-dnspod`

You need to create a `values.yaml` file to override `groupName` of the default value of the helm chart.

```yaml
groupName: <your group name>
```

```
helm install cert-manager-webhook-dnspod ./charts \
  --namespace cert-manager \
  -f values.yaml
```

### Issuer

Create a production issuer. And you could create a staging letsencrypt issuer if necessary.

```yaml
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
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
          groupName: <your group name>
          solverName: dnspod
          config:
            apiID: <your dnspod api id>
            apiTokenSecretRef:
              key: api-token
              name: dnspod-credentials
```

### Certificate

```yaml
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  # You could replace this name to your own
  # Pick any name as you wish
  name: wildcard-yourdomain-com # for *.yourdomain.com
spec:
  # Pick any name as you wish
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

****

> For contributors

## Development

Before you can run the test suite, you need to download the test binaries:

```sh
wget -O- https://storage.googleapis.com/kubebuilder-tools/kubebuilder-tools-1.14.1-darwin-amd64.tar.gz | tar x -
```

Then rename `testdata/my-custom-solver.example` as `testdata/my-custom-solver` and fulfill the values of DNSPod appId (`<your-dnspod-api-id>`) and apiToken (`<your-dnspod-api-token-base64>`).

Now we could run tests in debug mode with dlv

```sh
# You should change GROUP_NAME and TEST_ZONE_NAME to your own ones
GROUP_NAME=yourdomain.com \
TEST_ZONE_NAME=yourdomain.com. \
dlv test . -- -test.v
```

Or just run tests

```sh
GROUP_NAME=yourdomain.com \
TEST_ZONE_NAME=yourdomain. \
go test -v
```

## Contribution

This repo is essentially a fork of [jetstack/cert-manager-webhook-example](https://github.com/jetstack/cert-manager-webhook-example), so before you contribute to this repo, you could check the example.
