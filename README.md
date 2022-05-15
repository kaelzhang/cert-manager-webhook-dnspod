# Cert-Manager ACME webhook for DNSPod

Cert-manager webhook for DNSPod is a ACME webhook for [cert-manager](https://cert-manager.io) allowing users to use [DNSPod](https://www.dnspod.cn) for DNS01 challenge.

This is a **permanent** fork of [qqshfox/cert-manager-webhook-dnspod](https://github.com/qqshfox/cert-manager-webhook-dnspod) which is lack of maintainence.

Features
- Updated to cert-manager 1.1.0
- Updated to client-go 0.19.4
- No hardcoding in helm chart

Tested on production environment of
- Kubernetes 1.18.3

## Prerequisites

- A DNSPod [APP ID and API Token](https://support.dnspod.cn/Kb/showarticle/tsid/227/)
- A valid domain configured on DNSPod
- A Kubernetes cluster (v1.18+ recommended)
- Have [cert-manager](https://github.com/jetstack/cert-manager): >= 1.1.0 [installed](https://cert-manager.io/docs/installation/kubernetes/) within your kubernetes cluster.
- [Helm 3 installed](https://helm.sh/docs/intro/install/) on your local computer

## Installation

### Prepare for DNSPod

Create secret to store the API Token

```sh
kubectl --namespace cert-manager create secret generic \
  dnspod-credentials --from-literal=api-token='<DNSPOD_API_TOKEN>'
```

### Install `cert-manager-webhook-dnspod`

Clone this repository:

```
git clone https://github.com/kaelzhang/cert-manager-webhook-dnspod.git
```

You need to create a `values.yaml` file to override the default value of `groupName` for the helm chart.

```yaml
# The `groupName` here should be same as the value in cluster issuer below
groupName: <your group name>
```

```
helm install cert-manager-webhook-dnspod ./charts \
  --namespace cert-manager \
  -f values.yaml
```

### Issuer

Create a production issuer (And you could create a staging letsencrypt issuer instead if necessary)

Create a `cluster-issuer.yaml` file with the following content:

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

And run:

```
kubectl create -f cluster-issuer.yaml
```

### Certificate

#### Use Ingress to create the Certificate resource (Recommended)

A common use-case for cert-manager is requesting TLS signed certificates to secure your ingress resources.

This can be done by simply adding annotations to your Ingress resources and cert-manager will facilitate creating the Certificate resource for you without your concern. A small sub-component of cert-manager, ingress-shim, is responsible for this.

For details, see [here](https://cert-manager.io/docs/usage/ingress/)

Create a `ingress.yaml` file with the following content:

```yaml
apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  name: demo-ingress
  namespace: default
  annotations:
    # Should be the same as metadata.name of the cluster issuer
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
spec:
  tls:
  - hosts:
    - 'example.com'
    # Pick any name as you wish
    secretName: example-com-tls
  rules:
  - host: example.com
    http:
      paths:
      - path: /
        backend:
          serviceName: backend-service
          servicePort: 80
```

And run:

```
kubectl create -f ingress.yaml
```

#### Define the Certificate resource explicitly (Alternative)

If you don't use Ingress, you could define the certificate resource your own

Create a `certificate.yaml`:

```yaml
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  # You could replace this name to your own
  # Pick any name as you wish
  name: example-com # for example.com
spec:
  # Pick any name as you wish
  secretName: example-com-tls
  renewBefore: 240h
  dnsNames:
    - 'example.com'
  issuerRef:
    # The cluster issuer defined above
    name: letsencrypt-prod
    kind: ClusterIssuer
```

And run:

```
kubectl create -f certificate.yaml
```

### Check the result:

If the certificate is ready, you could see the following result:

```
$ kubectl get certificate

NAME          READY  SECRET           AGE
example-com   True   example-com-tls  2m1s
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
TEST_ZONE_NAME=yourdomain.com. \
go test -v
```
