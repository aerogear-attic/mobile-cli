## The Mobile Cli is a standalone cli that can also be used a kubectl / oc plugin

### Install

### Pre req

- have a local kubernetes or oc cluster via something like minikube or oc cluster up

### Setup the Custom Resource Definition

```
oc login -u system:admin
oc create -f artifacts/mobileclient_crd.yaml

```

In OpenShift, add the following to the edit and admin roles

``` 
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
You can do this via the edit command in oc
```
oc edit clusterrole admin # add the above and save
oc edit clusterrole edit # add the above and save
```

### Setup the plugin

- have the [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) command line tool or the [oc command line tool](https://docs.openshift.org/latest/cli_reference/get_started_cli.html#installing-the-cli) already installed
- It should be version k8s version 1.8 or OpenShift 3.7 or later
- create a new dir ```mkdir -p ~/.kube/plugins/mobile```  
- copy the plugin.yaml to the dir ```cp ./artifacts/cli_plugin.yaml ~/.kube/plugins/mobile```
- install the mobile cli binary onto your $PATH

### Basic usage

To access the plugin you can use the following command:
```kubectl plugin mobile <command>```

You can also use it standalone as it will pick up on your kube configuration:

```mobile --namespace=mine <command>``` notice we pass the namespace here.

**Passing flags**

To pass flags when using the plugin ensure to use the ```--``` option 

```kubectl plugin mobile -- create client --<someflag> ```


## Design
The design of the cli API attempts to give a familiar feel to users familiar wil the kubectl and oc CLIs. 
It is also intended to use parlance familiar to mobile developers in order to help them become more productive
and avoid needing to know the innards of various kubernetes resources.

### Command Structure

get 
    
    - clients
        - get clients #returns a list of mobile clients within the current namespace
        - get client <clientName> # returns a single mobile client
    - serviceconfig
        - get serviceconfigs # returns a list of mobile aware service configurations (this is the full unfiltered configuration data)
        - get serviceconfig <serviceName> #returns a single mobile aware service configuration (this is the full unfiltered configuration data) 
    - clientconfig
        - get clientconfig # returns the filtered configuration as should be consumed by mobile clients making use of the sdks 
    - clientbuild
        - get clientbuilds #returns a list of mobile client builds
        - get clientbuild <clientBuildName> #returns a single mobile client build
    - bindings   
        - get bindings # returns a list of bindings between mobile aware services and their consumers
        - get binding <bindingName> # return a single binding and the name of the services consuming it
    
    
create 
    
    - client
        - create client <clientName> <clientType> # will create a representation of the mobile client application
    - serviceconfig
        - create serviceconfig # this command will likely prompt for needed imputs    
    - binding
        - create binding <consuming_service> <bindable_service> # will create the binding and pod preset and optionally redeploy the integration_taget
        for example: mobile create binding fh-sync keycloak 
    - clientbuild
        - create clientbuild <clientID> <git_Source> [buildName] # will create a Jenkins PipeLine based buildconfig likely will need a several option flags to cover
        things like credentials and keys
    
delete

    - client
        - delete client <clientID> # removes the configmap or mobileclient object if we go with CRD
    - serviceconfig
        - delete serviceconfig <configName> # remove the configur        
    - binding
        - delete binding <consuming_servicet> <bindable_service> # removes all the objects created when the binding was enabled. 
        for example mobile delete binding fh-sync keycloak
    - clientbuild
        - delete the buildconfig and any other related object that back the client build.
        example: delete clientbuild <clientBuildName>    

replace


start

    - clientbuild <buildName>
    

stop

    - clientbuild <buildName>
                    