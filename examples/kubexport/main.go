package main

import (
	"context"
	"log/slog"

	"github.com/SAP/xp-clifford/cli"
	"github.com/SAP/xp-clifford/cli/export"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func exportLogic(_ context.Context, events export.EventHandler) error {
	slog.Info("export started")

	res := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"user":     "test-user",
			"password": "secret",
		},
	}
	events.Resource(res)

	events.Stop()
	return nil
}

func main() {
	cli.Configuration.ShortName = "kubexport"
	cli.Configuration.ObservedSystem = "Kubernetes"
	export.SetCommand(exportLogic)
	cli.Execute()
}
