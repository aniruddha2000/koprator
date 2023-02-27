# Koprator
----------

## Description

This is a Kubernetes operator using client-go. The operator subscribes to Kubernetes objects like pods & ConfigMaps. 
The ConfigMap that is in the Kube-system namespace has annotation data in it. The operator takes the annotation from 
the ConfigMap and adds those to every pod present in all namespaces. Secondly, when we add or delete the ConfigMap it 
sends a message to the Prometheus metrics server.

## Usage

### Start Kind cluster with kind config

```shell
$ kind create cluster --name koprator-test --config=kind-config.yaml
```

### Run in-cluster with helm

```shell
$ make helm
```

Access prometheus metrics server at [localhost:31000](http://localhost:31000)

### Run with kubeconfig from outside cluster

```shell
$ make build

$ make run
```

Access prometheus metrics server at [localhost:8080](http://localhost:8080)

## Debug

To `exec` into the pod -

```shell
$ kubectl debug -it <pod-name> --\image=ellerbrock/alpine-bash-curl-ssl -- /bin/bash
```