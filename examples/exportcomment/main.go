package main


import (
	"context"
	"log/slog"

	"github.com/SAP/xp-clifford/cli"
	"github.com/SAP/xp-clifford/cli/export"
	"github.com/SAP/xp-clifford/yaml"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func exportLogic(_ context.Context, events export.EventHandler) error {
	slog.Info("export command invoked")

	res := &unstructured.Unstructured{
	  Object: map[string]interface{}{
	      "user": "test-user-commented",
	      "password": "secret",
	  },
	}

	commentedResource := yaml.NewResourceWithComment(res)
	commentedResource.SetComment("don't deploy it, this is a test resource!")
	events.Resource(commentedResource)

	events.Stop()
	return nil
}

func main() {
	cli.Configuration.ShortName = "test"
	cli.Configuration.ObservedSystem = "test system"
	export.SetCommand(exportLogic)
	cli.Execute()
}
