package main

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/SAP/xp-clifford/cli"
	"github.com/SAP/xp-clifford/cli/export"
	"github.com/SAP/xp-clifford/cli/widget"
)

func exportLogic(ctx context.Context, events export.EventHandler) error {
	defer events.Stop()
	deepTree := widget.NewTreeOption("very-long-example-for-an-option-name")
	for i := 60; i >= 7; i-- {
		deepTree = widget.NewTreeOption(fmt.Sprintf("%d-very-long-example-for-an-option-name", i), deepTree)
	}

	selected, err := widget.TreeInput(ctx, "Select options", []widget.TreeOption{
		widget.NewTreeOption("option-1"),
		widget.NewTreeOption("option-2",
			widget.NewTreeOption("option-3"),
			widget.NewTreeOption("option-4",
				widget.NewTreeOption("option-5"),
				widget.NewTreeOption("option-6"),
			),
		),
		deepTree,
		widget.NewTreeOption("option-46"),
	})
	if err != nil {
		return err
	}
	slog.Info("data acquired", "selected", selected)
	return nil
}

func main() {
	cli.Configuration.ShortName = "test"
	cli.Configuration.ObservedSystem = "test system"
	export.SetCommand(exportLogic)
	cli.Execute()
}
