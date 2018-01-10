#!/usr/bin/env bash

k8=$GOPATH/src/k8s.io
code_gen=${k8}/code-generator


if [ ! -d ${code_gen} ]; then
   mkdir -p ${k8} && cd $k8 && git clone git@github.com:kubernetes/code-generator.git
fi

cd ${code_gen} && git checkout release-1.8

cd ${code_gen}
./generate-internal-groups.sh client github.com/aerogear/mobile-cli/pkg/client/mobile github.com/aerogear/mobile-cli/pkg/apis github.com/aerogear/mobile-cli/pkg/apis  "mobile:v1alpha1"
./generate-internal-groups.sh client github.com/aerogear/mobile-cli/pkg/client/servicecatalog github.com/aerogear/mobile-cli/pkg/apis github.com/aerogear/mobile-cli/pkg/apis  "servicecatalog:v1beta1"
