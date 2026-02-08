package main

import (
	"context"
	"errors"
	"log/slog"

	"github.com/SAP/xp-clifford/cli"
	"github.com/SAP/xp-clifford/cli/export"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func exportLogic(_ context.Context, events export.EventHandler) error {
	slog.Info("export command invoked")

	events.Warn(errors.New("generating test resource"))

	res := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"user":     "test-user-with-warning",
			"password": "secret",
		},
	}
	events.Resource(res)

	events.Stop()
	return nil
}

func main() {
	cli.Configuration.ShortName = "test"
	cli.Configuration.ObservedSystem = "test system"
	export.SetCommand(exportLogic)
	cli.Execute()
}
