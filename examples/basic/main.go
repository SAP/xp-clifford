package main

import (
	"github.com/SAP/xp-clifford/cli"
	_ "github.com/SAP/xp-clifford/cli/export"
)

func main() {
	cli.Configuration.ShortName = "test"
	cli.Configuration.ObservedSystem = "test system"
	cli.Execute()
}
