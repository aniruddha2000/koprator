# Koprator
----------

## Description

This is a Kubernetes operator using client-go. The operator subscribes to Kubernetes objects like pods & ConfigMaps. 
The ConfigMap that is in the Kube-system namespace has annotation data in it. The operator takes the annotation from 
the ConfigMap and adds those to every pod present in all namespaces. Secondly, when we add or delete the ConfigMap it 
sends a message to the Prometheus metrics server.

## Usage

```shell
$ make build

$ make run
```