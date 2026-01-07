package cli_test

import "github.com/SAP/xp-clifford/cli"

func ExampleExecute() {
	cli.Configuration.ShortName = "ts"
	cli.Configuration.ObservedSystem = "test system"
	cli.Execute()
}
