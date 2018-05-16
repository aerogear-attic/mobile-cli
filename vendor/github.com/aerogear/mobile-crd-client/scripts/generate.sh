#!/usr/bin/env bash

k8=$GOPATH/src/k8s.io
code_gen=${k8}/code-generator
apimachinery=${k8}/apimachinery
kubernetes=${k8}/kubernetes

if [ ! -d ${code_gen} ]; then
   mkdir -p ${k8} && cd $k8 && git clone git@github.com:kubernetes/code-generator.git
fi

if [ ! -d ${apimachinery} ]; then
  cd ${k8} && git clone git@github.com:kubernetes/apimachinery.git && cd ${apimachinery} && git checkout kubernetes-1.8.0
fi

if [ ! -d ${kubernetes} ]; then
   cd $k8 && git clone --depth=1 git@github.com:kubernetes/kubernetes
fi

cd ${code_gen} && git checkout release-1.8

cd ${code_gen}
./generate-internal-groups.sh client github.com/aerogear/mobile-crd-client/pkg/client/mobile github.com/aerogear/mobile-crd-client/pkg/apis github.com/aerogear/mobile-crd-client/pkg/apis  "mobile:v1alpha1"
./generate-internal-groups.sh client github.com/aerogear/mobile-crd-client/pkg/client/servicecatalog github.com/aerogear/mobile-crd-client/pkg/apis github.com/aerogear/mobile-crd-client/pkg/apis  "servicecatalog:v1beta1"
