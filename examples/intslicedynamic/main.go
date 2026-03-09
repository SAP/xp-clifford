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
		"secure", secureParam.Value(),
		"secure-is-set", secureParam.IsSet(),
		"ports", portsParam.Value(),
		"num-of-ports", len(portsParam.Value()),
		"ports-is-set", portsParam.IsSet(),
	)

	ports, err := portsParam.ValueOrAsk(ctx)
	if err != nil {
		return err
	}

	slog.Info("data acquired", "ports", ports)

	events.Stop()
	return nil
}

func possibleProtocols() ([]int, error) {
	if secureParam.Value() {
		return []int{443, 22}, nil
	}
	return []int{23, 80}, nil
}

var secureParam = configparam.Bool("secure", "secure protocol").
	WithShortName("s").
	WithEnvVarName("SECURE")

var portsParam = configparam.IntSlice("ports", "list of supported ports").
	WithShortName("p").
	WithPossibleValuesFn(possibleProtocols)

func main() {
	cli.Configuration.ShortName = "test"
	cli.Configuration.ObservedSystem = "test system"
	export.AddConfigParams(secureParam, portsParam)
	export.SetCommand(exportLogic)
	cli.Execute()
}
