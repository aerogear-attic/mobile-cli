# Adding a new CLI command

The mobile CLI uses the [cobra](https://github.com/spf13/cobra) library to provide a consistent framework for building out the CLI tool.


## Adding a new base command
### What is a base command
A base command is a command that works with an entirely new resource or that adds a new verb i.e (get, delete)
### Changes required to add a new base command
To add a new base command that operates on a new resource type you should add a new file under the ```pkg/cmd``` directory
named ```<resourceName>.go``` with an accompanying test file ```<resourceName>_test.go```.
Once this is done, then you should follow the patterns in the other base commands:

- Create a new type  ```<resourceName>Cmd```
- Create new constructor ```New<resourceName>Cmd``` 
- Dependencies such as the kubernetes clients should be passed to this constructor as their interface types to allow for simpler testing
- The cobra commands should then be returned from methods from this the base type. See below:

```go
type MyResourceCMD struct{}

func (mr *MyResourceCMD) ListMyResourceCMD() *cobra.Command{}
```

This command should then be wired up in ```main.go``` inside the ```cmd/mobile``` pkg. 

If it is being added to an existing verb command then add this command in the same way as the other commands. If it is a new verb command then you will want to create a new block of it and its sub commands.

Example of adding a new resource command:

```go
var (
	out              = os.Stdout
	rootCmd          = cmd.NewRootCmd()
	clientCmd        = cmd.NewClientCmd(mobileClient, out)
	bindCmd          = cmd.NewIntegrationCmd(scClient, k8Client)
	serviceConfigCmd = cmd.NewServiceConfigCommand(k8Client)
	clientCfgCmd     = cmd.NewClientConfigCmd(k8Client)
	clientBuilds     = cmd.NewClientBuildsCmd()
	svcCmd           = cmd.NewServicesCmd(scClient, k8Client, out)
	// new command added here
	myResource       = cmd.NewMyResourceCmd(scClient,k8Client, out)
)	
	
// create
{
	createCmd := cmd.NewCreateCommand()
	createCmd.AddCommand(svcCmd.CreateServiceInstanceCmd())
	createCmd.AddCommand(bindCmd.CreateIntegrationCmd())
	createCmd.AddCommand(clientCmd.CreateClientCmd())
	createCmd.AddCommand(serviceConfigCmd.CreateServiceConfigCmd())
	createCmd.AddCommand(clientBuilds.CreateClientBuildsCmd())
	// sub command added here
	createCmd.AddCommand(myResource.CreateMyResourceCmd())
	
	rootCmd.AddCommand(createCmd)
}
```


Example of adding a new verb command:


```go
// twist
{
	twistCmd := cmd.NewTwistCmd()
	twistCmd.AddCommand(myResource.TwistMyResourceCmd())
	
	// important to add it to the root command
	rootCmd.AddCommand(twistCmd)
}

```

### Adding a new command to an existing resource type

You would do this when adding additional functionality to an existing resource type:

- First add a new method to the resource type under <resourceName.go>

```go
func (cbc *ClientBuildsCmd) ListClientBuildsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clientbuild",
		Short: "get a specific clientbuild for a mobile client",
		RunE: func(cmd *cobra.Command, args []string) error {
			...
		},
	}
	return cmd
}
```

You should also then add the testcase for this command:

```go
func TestClientBuildsCmd_ListClientBuildsCmd(t *testing.T) {
...
}
```