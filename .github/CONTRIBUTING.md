# Contributing to the mobile CLI

This document explains how to set up a development environment and get involved with Mobile CLI project.

Before anything else, [fork](https://help.github.com/articles/fork-a-repo/) the Mobile CLI project.

## Tools We use

* Git
* Go
* Make

## Set Up OpenShift

The Mobile CLI targets Kubernetes and is intended to be developed against a running Kubernetes cluster,
we use OpenShift as our Kubernetes distribution. The Mobile CLI is intended to help you work with Mobile Services running ontop of OpenShift.
To provision these services, we levarage the [Service Catalog](https://github.com/kubernetes-incubator/service-catalog) and the [Ansible Service Broker](https://github.com/openshift/ansible-service-broker).
To help get this infrastructure set up there is a [ansible based installer](https://github.com/aerogear/mobile-core#installing-from-a-development-release) provided.

## Clone the repositiory

As we are using Go, the path you clone this repo into is important.

* create the directory `mkdir -p $GOPATH/src/github.com/aerogear`
* clone the repo `cd $GOPATH/src/github.com/aerogear && git clone git@github.com:aerogear/mobile-CLI.git`
* add your own fork as the upstream target `git add remote upstream <your fork>`

## Building the Mobile CLI

To build the CLI locally you can run `make build` this command will run a set of checks and the unit tests before compiling the binary and outputting it into the current directory,
if you only want to build the binary itself you can simply run `make build_binary`.
Once built, you can access this binary and use it from the command line `./mobile`
Remember however that the CLI is intended as a Kubernetes plugin so expects to find kube configuration in `~/.kube/config`. If you have setup OpenShift this should
already be in place.

## Submitting changes to the Mobile CLI

### Before making a pull request

There are a few things you should keep in mind before creating a PR.

* New code should have corresponding unit tests. An example of how we approach unit testing can be found in [clients_test.go](https://github.com/aerogear/mobile-CLI/blob/master/pkg/cmd/clients_test.go).

* Ensure for new commands to read the [adding a new command](https://github.com/aerogear/mobile-CLI/doc/adding_new_cmd.md) doc before hand.

* You must run ```make build``` before creating the PR and ensure it must execute with no errors.

* When needed, provide an explanation of the command and the expected output to help others that may review and test the change.

### Making a pull request

Make a [pull request (PR)](https://help.github.com/articles/using-pull-requests) in the standard way.

Use [WIP] at the beginning of the title (ie. [WIP] Add feature to the CLI) to mark a PR as a Work in Progress.

If you are not a member of the [aerogear org](https://github.com/aerogear), the build job will pause for approval from a trusted approver.
Anyone who can login to Jenkins can approve.

Your PR will then be reviewed, questions may be asked and changes requested.

Upon successful review, someone will approve the PR in the review thread. Depending on the size of the change, we may wait for 2 LGTM from reviewers before merging.


### Major features

The aerogear community uses a proposal process when introducing a major feature in order to encourage collaboration and building the best solution.

Major features are things that take about 2 weeks of development or introduce disruptive changes to the code base.

Start the proposal process by reviewing the [proposal template](https://github.com/aerogear/proposals/blob/master/template.md). Use this document to guide how to write a proposal. Then, submit it as a pull request where the community will review the plan.

The proposal process also requires two approvals from the community before merging.

## Stay in touch

* IRC: Join the conversation on Freenode: #aerogear
* Email: Subscribe to the [aerogear mailing list](https://lists.jboss.org/mailman/listinfo/aerogear-dev)
