# mobile-crd-client

This repo contains Golang client libraries for the following:

* The Mobile Client Custom Resource Definition for Kubernetes
* The Kubernetes Service Catalog

The client libraries were generated using [kubernetes/code-generator](https://github.com/kubernetes/code-generator).

## How to Import

The following demonstrates how to import the client libraries contained in this repository.

```
import (
	mobile "github.com/aerogear/mobile-crd-client/pkg/client/mobile/clientset/versioned"
	servicecatalog "github.com/aerogear/mobile-crd-client/pkg/client/servicecatalog/clientset/versioned"
)
```

## How to Generate

The libraries were generated using `scripts/generate.sh`.

```
$ cd scripts && ./generate.sh
Generating clientset for mobile:v1alpha1 at github.com/aerogear/mobile-crd-client/pkg/client/mobile/clientset
Generating clientset for servicecatalog:v1beta1 at github.com/aerogear/mobile-crd-client/pkg/client/servicecatalog/clientset
```