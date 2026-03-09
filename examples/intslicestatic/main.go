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

	ports, err := testParam.ValueOrAsk(ctx)
	if err != nil {
		return err
	}

	slog.Info("data acquired", "ports", ports)

	events.Stop()
	return nil
}

var testParam = configparam.IntSlice("ports", "list of supported ports").
	WithShortName("p").
	WithPossibleValues([]int{22, 23, 80, 443})

func main() {
	cli.Configuration.ShortName = "test"
	cli.Configuration.ObservedSystem = "test system"
	export.AddConfigParams(testParam)
	export.SetCommand(exportLogic)
	cli.Execute()
}
