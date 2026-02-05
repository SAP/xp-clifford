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
		"port", testParam.Value(),
		"is-set", testParam.IsSet(),
	)

	// If not set, ask the value
	port, err := testParam.ValueOrAsk(ctx)
	if err != nil {
		return err
	}

	slog.Info("value set by user", "value", port)

	events.Stop()
	return nil
}

var testParam = configparam.Int("port", "port number").
	WithShortName("p").
	WithEnvVarName("PORT").
	WithDefaultValue(443)

func main() {
	cli.Configuration.ShortName = "test"
	cli.Configuration.ObservedSystem = "test system"
	export.AddConfigParams(testParam)
	export.SetCommand(exportLogic)
        cli.Execute()
}
