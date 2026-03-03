package main

import (
	"context"
	"log/slog"

	"github.com/SAP/xp-clifford/cli"
	"github.com/SAP/xp-clifford/cli/export"
)

func exportLogic(_ context.Context, events export.EventHandler) error {
	slog.Debug("export command invoked")
	events.Stop()
	return nil
}

func main() {
	cli.Configuration.ShortName = "test"
	cli.Configuration.ObservedSystem = "test system"
	export.SetCommand(exportLogic)
	cli.Execute()
}
