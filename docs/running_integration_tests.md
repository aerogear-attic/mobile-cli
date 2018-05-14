# Running integration tests

Our integration tests execute `mobile` commands which are ran against local mobile enabled Openshift cluster.

## Prerequisites
- mobile enabled Openshift cluster running on the host where you want to execute the tests [local setup](https://github.com/aerogear/mobile-core/blob/master/docs/walkthroughs/local-setup.adoc)

## Execution
run `make integration` - this will build `mobile` binary and use it when executing commands
