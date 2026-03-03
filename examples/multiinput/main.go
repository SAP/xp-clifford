package main

import (
	"context"
	"log/slog"

	"github.com/SAP/xp-clifford/cli"
	"github.com/SAP/xp-clifford/cli/export"
	"github.com/SAP/xp-clifford/cli/widget"
)

func exportLogic(ctx context.Context, events export.EventHandler) error {
	slog.Info("export command invoked")

	protocols, err := widget.MultiInput(ctx,
		"Select the supported protocols",
		[]string{
			"FTP",
			"HTTP",
			"HTTPS",
			"SFTP",
			"SSH",
		},
	)

	slog.Info("data acquired",
		"protocols", protocols,
	)

	events.Stop()
	return err
}

func main() {
	cli.Configuration.ShortName = "test"
	cli.Configuration.ObservedSystem = "test system"
	export.SetCommand(exportLogic)
	cli.Execute()
}
