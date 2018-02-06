package integration

import (
	"flag"
)

var namespace = flag.String("namespace", "", "Openshift namespace (most often Project) to run our integration tests in")
var goldenFiles = flag.String("goldenFiles", "", "Path to folder containing folders with goldenfiles. If not set, will assume the local ./integration folder")
var prefix = flag.String("prefix", "test", "Client name to be created")
var executable = flag.String("executable", "", "Executable under test")
var update = flag.Bool("update", false, "update golden files")
