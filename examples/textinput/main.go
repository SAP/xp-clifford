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

	username, err := widget.TextInput(ctx, "Username", "anonymous", false)
	if err != nil {
		return err
	}

	password, err := widget.TextInput(ctx, "Password", "", true)
	if err != nil {
		return err
	}

	slog.Info("data acquired",
		"username", username,
		"password", password,
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
