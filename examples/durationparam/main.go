package main

import (
	"context"
	"log/slog"
	"time"

	"github.com/SAP/xp-clifford/cli"
	"github.com/SAP/xp-clifford/cli/configparam"
	"github.com/SAP/xp-clifford/cli/export"
)

func exportLogic(ctx context.Context, events export.EventHandler) error {
	slog.Info("export command invoked",
		"timeout", testParam.Value(),
		"is-set", testParam.IsSet(),
	)

	// If not set, ask the value
	timeout, err := testParam.ValueOrAsk(ctx)
	if err != nil {
		return err
	}

	slog.Info("value set by user", "value", timeout)

	events.Stop()
	return nil
}

var testParam = configparam.Duration("timeout", "request timeout").
	WithShortName("t").
	WithEnvVarName("TIMEOUT").
	WithDefaultValue(30*time.Second)

func main() {
	cli.Configuration.ShortName = "test"
	cli.Configuration.ObservedSystem = "test system"
	export.AddConfigParams(testParam)
	export.SetCommand(exportLogic)
        cli.Execute()
}
