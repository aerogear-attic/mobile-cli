## Service Configuration

### Overview
the Service Config stores information about a mobile service and is backed by a secret. This information is then used to populate your mobile client's config.

This information could be anything but often is made up of values such as the URL of the service and perhaps some headers and configuration particular to that service.

### Retrieving Configurations
All service configs for a particular namespace can be retrieved with the following command:
```sh
mobile get clientconfig example_client_id --namespace=myproject
```

Which will produce output like the following:
```sh
+----------------+-------------------+----------------+-------------------------------------------------------+
|       ID       |      NAME         |      TYPE      |                          URL                          |
+----------------+-------------------+----------------+-------------------------------------------------------+
| Client ID      | example_client_id |                |                                                       |
| fh-sync-server | fh-sync-server    | fh-sync-server | https://fh-sync-server-myproject.192.168.64.74.nip.io |
| keycloak       | keycloak          | keycloak       | https://keycloak-myproject.192.168.64.74.nip.io       |
| prometheus     | prometheus        | prometheus     | https://prometheus-myproject.192.168.64.74.nip.io     |
+----------------+-------------------+----------------+-------------------------------------------------------+
```

For more verbose config details of each service, you can request the output in JSON format instead, with the following command:
```sh
mobile get clientconfig example_client_id --namespace=myproject -o json
```

Which will produce output similar to the following (newlines and indentation have been added for readability):
```
{
  "version": "1.0",
  "cluster_name": "192.168.64.74:8443",
  "namespace": "myproject",
  "client_id": "example_client_id",
  "services": [
    {
      "id": "fh-sync-server",
      "name": "fh-sync-server",
      "type": "fh-sync-server",
      "url": "https://fh-sync-server-myproject.192.168.64.74.nip.io",
      "config": {
        "url": "https://fh-sync-server-myproject.192.168.64.74.nip.io"
      }
    },
    {
      "id": "keycloak",
      "name": "keycloak",
      "type": "keycloak",
      "url": "https://keycloak-myproject.192.168.64.74.nip.io",
      "config": {
        "auth-server-url": "https://keycloak-myproject.192.168.64.74.nip.io/auth",
        "clientId": "juYAlRlhTyYYmOyszFa",
        "realm": "myproject",
        "resource": "juYAlRlhTyYYmOyszFa",
        "ssl-required": "external",
        "url": "https://keycloak-myproject.192.168.64.74.nip.io/auth"
      }
    },
    {
      "id": "prometheus",
      "name": "prometheus",
      "type": "prometheus",
      "url": "https://prometheus-myproject.192.168.64.74.nip.io",
      "config": {}
    }
  ]
}
```

### Understanding the JSON format
Firstly, the parent object in the JSON output is described below:

#### version
The version of the JSON structure used in this response.

#### cluster_name
An identifier of the cluster this config was retrieved from.

#### namespace
The namespace these configs were retrieved from.

#### client_id
The client id of the mobile application using these configs.

#### services
An array of configuration values for the provisioned mobile services in this namespace.

### The service object
The service object contains specific configuration values for each mobile service provisioned to a specific namespace.

Each service has a common set of values:
#### id
A unique identifier of this specific deployment of this service.

#### name
A human-readable identifier of this service.

#### type
A way of categorising services, e.g. Authentication, Storage, etc...

#### url
The URL that this service can be reach at

#### config
The config is a loosely defined object where any extra details specific to a particular service that may be required to make proper use of this service will be stored.

For example, in the KeyCloak service in the snippet above, the config contains the name of the realm, clientID and other details that would be required to make use of KeyCloak from a client.