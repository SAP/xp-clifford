package main

import (
	"context"
	"log/slog"

	"github.com/SAP/xp-clifford/cli"
	"github.com/SAP/xp-clifford/cli/configparam"
	_ "github.com/SAP/xp-clifford/cli/export"
)

func login(_ context.Context) error {
	slog.Info("login invoked")
	return nil
}

func main() {
	cli.Configuration.ShortName = "test"
	cli.Configuration.ObservedSystem = "test system"

	var loginSubCommand = &cli.BasicSubCommand{
		Name:         "login",
		Short:        "Login demo subcommand",
		Long:         "A subcommand demonstrating xp-clifford capabilities",
		ConfigParams: []configparam.ConfigParam{},
		Run:          login,
	}

	cli.RegisterSubCommand(loginSubCommand)

	cli.Execute()
}
