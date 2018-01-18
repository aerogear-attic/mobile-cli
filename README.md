## The Mobile CLI is a standalone CLI that can also be used a kubectl / oc plugin

## Note this is still under construction and not yet fit for use

The mobile CLI focuses on a small set of commands to empower mobile focused developers to consume and take full advantage of the RedHat mobile suite
of services ontop of Kubernetes / OpenShift. 

It uses a language familiar to mobile developers and abstracts away some of the complexity of dealing with Kubernetes / OpenShift which can be 
initially daunting and overwhelming.

### Examples
Note not all of these commands currently exist but are present below to show the general concept

```
mobile get services

mobile provision fh-sync

mobile --namespace=myproject get clientconfig

mobile create integration fh-sync keycloak

mobile create clientbuild <MobileClientID> <Git_Source_Url> [buildName]

mobile get buildartifact <clientBuildID> 
``` 

### Checkout 

```
mkdir -p $GOPATH/src/github.com/aerogear
cd $GOPATH/src/github.com/aerogear
git clone https://github.com/aerogear/mobile-cli
```

### Build 

```
glide install
make build
```

To test, run:

```
./mobile
```

### Build for APB usage

To use the mobile CLI inside an APB container it needs to be compiled for the linux/amd64 platform:

```
make build_linux
```

### Install

### Pre req

- Have a local kubernetes or oc cluster via something like minikube or oc cluster up
- Install glide (https://github.com/Masterminds/glide), e.g. `brew install glide`

### Setup the Custom Resource Definition

```
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

### Setup the plugin

- have the [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) command line tool or the [oc command line tool](https://docs.openshift.org/latest/cli_reference/get_started_cli.html#installing-the-cli) already installed
- it should be version k8s version 1.8 or OpenShift 3.7 or later
- create a new dir ```mkdir -p ~/.kube/plugins/mobile```  
- copy the plugin.yaml to the dir ```cp ./artifacts/cli_plugin.yaml ~/.kube/plugins/mobile```
- install the mobile CLI binary onto your $PATH

### Basic usage

To access the plugin you can use the following command:
```kubectl plugin mobile <command>```

You can also use it standalone as it will pick up on your kube configuration:
```mobile --namespace=mine <command>``` notice we pass the namespace here.

**Passing flags**

To pass flags when using the plugin ensure to use the ```--``` option:
```kubectl plugin mobile -- create client --<someflag> ```


## Design

The design of the CLI API attempts to give a familiar feel to users familiar wil the kubectl and oc CLIs. 
It is also intended to use parlance familiar to mobile developers in order to help them become more productive
and avoid needing to know the innards of various kubernetes resources.

## Core Objects or Resources

In a similar fashion to the oc and kubectl CLI, we have some core resources that we work with. Some of these are backed by things like secrets while others are
defined as custom resources.

- **MobileClient:** the mobile client is a resource that represents your mobile client app as part of the OpenShift UI. It gives us the context and information needed to show you relevant information
around your particular mobile runtime as well as allowing us setup the different kind of client builds required.

- **ClientConfig** the client config, is a resource created by aggregating together all of the available service configs. This resource is configuration
required in order to consume your mobile aware services from your mobile client. It is used by the client SDKs for the various mobile services.

-  **ServiceConfig:** the Service Config stores information about a mobile aware service and is backed by a secret. This information is then used to populate your mobile client's config.
This information could be anything but often is made up of values such as the URI of the service and perhaps some headers and configuration particular to that service.

- **ClientBuild** the ClientBuild is backed by a regular BuildConfig however the CLI will help you create this BuildConfig with as little effort as possible. Allowing you to focus on
just the mobile parts rather than needing to understand how to setup and manage a buildconfig and builds. For example, it will help you manage build credentials, and keys and ensure the build integrates
seamlessly with the areogear mobile build farm .

- **Binding** the Binding is backed by a binding resource in the service catalog. Once again we try to remove the need to understand how to create the
native objects so that you can focus on being productive and building your mobile app. When doing a binding, you will be able to integrate different
mobile services together. For example when using sync and keycloak you can bind them together and have your sync service protected by keycloak. This is as simple as
```mobile bind fh-sync keycloak```

### Command Structure

get 
    
    - clients
        - get clients #returns a list of mobile clients within the current namespace
        - get client <clientName> # returns a single mobile client
    - serviceinstances
        - get serviceinstances <serviceName> # Returns a list of provisioned serviceInstances based on the service name.
    - services
        - get services # Returns mobile aware services that can be provisioned to your namespace
    - serviceconfig
        - get serviceconfigs # returns a list of mobile aware service configurations (this is the full unfiltered configuration data)
        - get serviceconfig <serviceName> #returns a single mobile aware service configuration (this is the full unfiltered configuration data) 
    - clientconfig
        - get clientconfig # returns the filtered configuration as should be consumed by mobile clients making use of the sdks 
    - clientbuild
        - get clientbuilds #returns a list of mobile client builds
        - get clientbuild <clientBuildName> #returns a single mobile client build
    - integrations   
        - get integrations # returns a list of integrations between mobile aware services and their consumers
        - get integration <name> # return a single integration and the name of the services consuming it
    
    
create 
    
    - client
        - create client <clientName> <clientType> # will create a representation of the mobile client application
    - serviceinstance
        - create serviceinstance <serviceName> # this command will likely prompt for needed imputs
    - serviceconfig
        - create serviceconfig # this command will likely prompt for needed imputs    
    - integration
        - create integration <consuming_service_inst_id> <providing_service_inst_id> # will create the binding and pod preset and optionally redeploy the consuming service
        for example: mobile create integration fh-sync keycloak 
    - clientbuild
        - create clientbuild <clientID> <git_Source> [buildName] # will create a Jenkins PipeLine based buildconfig likely will need a several option flags to cover
        things like credentials and keys
    
delete

    - client
        - delete client <clientID> # removes the configmap or mobileclient object if we go with CRD
    - serviceinstance
        - delete serviceinstance <serviceInstanceID> # deletes a service instance and other objects created when provisioning the services instance such as pod presets
    - serviceconfig
        - delete serviceconfig <configName> # remove the configur        
    - integration
        - delete integration <consuming_service_inst_id> <providing_service_inst_id> # removes all the objects created when the integratio was enabled. 
        for example mobile delete binding fh-sync keycloak
    - clientbuild
        - delete the buildconfig and any other related object that back the client build.
        example: delete clientbuild <clientBuildName>    

replace


start

    - clientbuild <buildName>
    

stop

    - clientbuild <buildName>
                    

## Contributing 

Check the [`CONTRIBUTING.md`](https://github.com/aerogear/mobile-cli/blob/master/.github/CONTRIBUTING.md) file. 