package main

import (
	"context"
	"log/slog"

	"github.com/SAP/xp-clifford/cli"
	"github.com/SAP/xp-clifford/cli/configparam"
	"github.com/SAP/xp-clifford/cli/export"
)

func exportLogic(ctx context.Context, events export.EventHandler) error {
	slog.Info("export command invoked",
		"ports", testParam.Value(),
		"num-of-ports", len(testParam.Value()),
		"is-set", testParam.IsSet(),
	)

	events.Stop()
	return nil
}

var testParam = configparam.IntSlice("port", "list of supported port numbers").
	WithShortName("p")

func main() {
	cli.Configuration.ShortName = "test"
	cli.Configuration.ObservedSystem = "test system"
	export.AddConfigParams(testParam)
	export.SetCommand(exportLogic)
	cli.Execute()
}
