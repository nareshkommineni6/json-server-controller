# json-server-controller

Kubernetes controller for managing json-server instances using CRDs.

## Implementation Checklist

- [x] Kubebuilder project with CRD for `JsonServer`
- [x] CRD fields: `replicas` and `jsonConfig`
- [x] Controller creates Deployment, Service, ConfigMap
- [x] Admission webhook validates name starts with `app-`
- [x] Admission webhook validates `jsonConfig` is valid JSON
- [x] Status updates: `Synced` or `Error` state
- [x] Support for `kubectl scale` command
- [x] CI/CD pipeline with ttl.sh for image publishing

## Prerequisites

- Go 1.21+
- Docker
- Kind/K3d/Minikube
- kubectl

## Setup

```bash
# create a cluster
kind create cluster

# install the CRD
make install

# run the controller locally
ENABLE_WEBHOOKS=false make run
```

## Usage

Create a JsonServer resource:

```yaml
apiVersion: example.com/v1
kind: JsonServer
metadata:
  name: app-my-server
spec:
  replicas: 2
  jsonConfig: |
    {
      "people": [
        {"id": 1, "name": "John"},
        {"id": 2, "name": "Jane"}
      ]
    }
```

Apply it:
```bash
kubectl apply -f config/samples/example_v1_jsonserver.yaml
```

This will create:
- A Deployment running `backplane/json-server`
- A Service exposing port 3000
- A ConfigMap with the json data

## Testing

```bash
# check status
kubectl get jsonservers

# port forward and test
kubectl port-forward svc/app-my-server 3000:3000
curl http://localhost:3000/people
```

## Validation

The admission webhook validates:
- Name must start with `app-`
- jsonConfig must be valid JSON

## Scaling

```bash
kubectl scale jsonserver app-my-server --replicas=3
```

## Building

```bash
# build image
make docker-build IMG=ttl.sh/json-server-controller:1h

# push
make docker-push IMG=ttl.sh/json-server-controller:1h

# deploy to cluster
make deploy IMG=ttl.sh/json-server-controller:1h
```

## Cleanup

```bash
kubectl delete jsonservers --all
make undeploy
make uninstall
kind delete cluster
```
