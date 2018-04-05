[![Go Report Card](https://goreportcard.com/badge/github.com/golang/crypto)](https://goreportcard.com/report/github.com/golang/crypto)

[![Coverage Status](https://coveralls.io/repos/github/aerogear/mobile-cli/badge.svg?branch=add-go-report-card-to-readme)](https://coveralls.io/github/aerogear/mobile-cli?branch=add-go-report-card-to-readme)

# Mobile CLI

**NOTE: The Mobile CLI is still under construction and not yet fit for use.**

The Mobile CLI is a standalone CLI that can also be used as a kubectl or oc plugin.

It focuses on a small set of commands to empower mobile focused developers to consume and take full advantage of the RedHat mobile suite of services ontop of Kubernetes/OpenShift. 

It uses a language familiar to mobile developers and abstracts away some of the complexity of dealing with Kubernetes/OpenShift which can be initially daunting and overwhelming.

## Examples
**Note: Not all of these commands currently exist but are present below to show the general concept.**

```bash
mobile get services
mobile create serviceinstance <serviceName> --namespace=<namespace>
mobile get clients --namespace=<namespace>
mobile get clientconfig <mobileClientID> --namespace=<namespace> 
mobile create integration <consumingServiceInstanceID> <providingServiceInstanceID> --namespace=<namespace>
``` 

## CLI Installation
### Pre-requisites
- Have a local Kubernetes or OpenShift cluster with mobile clients and services available via minikube, [mobile core installer](https://github.com/aerogear/mobile-core/blob/master/docs/walkthroughs/local-setup.adoc) or [minishift](https://github.com/aerogear/minishift-mobilecore-addon).
- Install [glide](https://github.com/Masterminds/glide)
- Install [go](https://golang.org/doc/install)

### Clone this repository

```bash
mkdir -p $GOPATH/src/github.com/aerogear
cd $GOPATH/src/github.com/aerogear
git clone https://github.com/aerogear/mobile-cli
```

### Build the CLI Binary

```bash
glide install
make build
```

To test, run:

```bash
./mobile
```

### Build for APB usage

To use the mobile CLI inside an APB container it needs to be compiled for the linux/amd64 platform:

```bash
make build_binary_linux
```

### Setup the Custom Resource Definition

```bash
oc login -u system:admin
oc create -f artifacts/mobileclient_crd.yaml
```

In OpenShift, add the following to the edit and admin roles:

```yml
- apiGroups:
  - mobile.k8s.io
  attributeRestrictions: null
  resources:
  - mobileclients
  verbs:
  - create
  - delete
  - get
  - list
  - patch
  - update
  - watch
```

You can do this via the edit command in oc:

```bash 
oc edit clusterrole admin # add the above and save
oc edit clusterrole edit # add the above and save
```

### Setup as Kubectl/OC plugin

- have the [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) command line tool or the [oc command line tool](https://docs.openshift.org/latest/cli_reference/get_started_cli.html#installing-the-cli) already installed
- it should be version k8s version 1.8 or OpenShift 3.7 or later
- create a new dir ```mkdir -p ~/.kube/plugins/mobile```  
- copy the cli_plugin.yaml file to the dir ```cp ./artifacts/cli_plugin.yaml ~/.kube/plugins/mobile```
and rename the ```cli_plugin.yaml``` file to ```plugin.yaml```
- install the mobile CLI binary onto your $PATH

## Basic usage

To use the mobile CLI as a plugin you can use the following command:
``` bash
kubectl plugin mobile <command>
oc plugin mobile <command>
```

**Passing flags**

To pass flags when using the mobile CLI as a plugin, ensure to use the ```--``` option:
```bash
kubectl plugin mobile create client -- --<someflag>
oc plugin mobile create client -- --<someflag>
```

The mobile CLI can also be use standalone as it will pick up on your kube configuration:
```bash
mobile <command> --namespace=mine 
``` 

**NOTE: When this CLI is used as an OC plugin, you do not need to provide the --namespace flag.**

## Design

The design of the CLI API attempts to give a familiar feel to users familiar with the kubectl and oc CLIs.  It is also intended to use parlance familiar to mobile developers in order to help them become more productive and avoid needing to know the innards of various kubernetes resources.

## Core Objects or Resources

In a similar fashion to the oc and kubectl CLI, we have some core resources that we work with. Some of these are backed by things like secrets while others are defined as custom resources.

- **MobileClient:** The mobile client is a resource that represents your mobile client application as part of the OpenShift UI. It gives us the context and information needed to show you relevant information around your particular mobile runtime as well as allowing us to setup the different kind of client builds required.

- **ClientConfig** The client config, is a resource created by aggregating together all of the available service configs. This resource is the configuration required in order to consume your mobile aware services from your mobile client. It is used by the client SDKs for the various mobile services.

-  **ServiceConfig:** The service config contains the services' information that is used to configure the Mobile SDK. For more information see [here](./doc/service_config.md).

- **ClientBuild** The client build is backed by a regular BuildConfig, however the CLI will help you create this BuildConfig with as little effort as possible. This allows you to focus on just the mobile parts rather than needing to understand how to setup and manage a buildconfig and builds. For example, it will help you manage build credentials, and keys and ensure the build integrates seamlessly with the aereogear mobile build farm.

- **Binding** The binding is backed by a binding resource in the service catalog. Once again we try to remove the need to understand how to create the native objects so that you can focus on being productive and building your mobile app. When doing a binding, you will be able to integrate different mobile services together. For example when using sync and keycloak you can bind them together and have your sync service protected by keycloak. This is as simple as
```mobile create integration <consuming_service_instance_id> <providing_service_instance_id>```

## Command Structure

### get
```
  client           gets a single mobile client in the namespace
  clients          gets a list of mobile clients represented in the namespace
  clientconfig     get clientconfig returns a client ready filtered configuration of the available services.
  integration      get a single integration
  integrations     get a list of the current integrations between services
  serviceconfig    get a mobile aware service definition
  serviceconfigs   get a list of deployed mobile enabled services
  serviceinstances get a list of provisioned service instances based on the service name.
  services         get mobile aware services that can be provisioned to your namespace
```
    
### create 
```
  client          create a mobile client representation in your namespace
  integration     integrate certain mobile services together. mobile get services will show you what can be integrated.
  serviceconfig   create a new service config
  serviceinstance create a running instance of the given service
```
    
### delete
```
  client          deletes a single mobile client in the namespace
  integration     delete the integration between mobile services.
  serviceconfig   delete a service config
  serviceinstance deletes a service instance and other objects created when provisioning the services instance, such as pod presets
```
                    
## Contributing 

Check the [`CONTRIBUTING.md`](https://github.com/aerogear/mobile-cli/blob/master/.github/CONTRIBUTING.md) file. 