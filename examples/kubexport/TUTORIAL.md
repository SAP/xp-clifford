# Getting Started with xp-clifford

In this tutorial, you will build `kubexport` — a CLI tool that exports Kubernetes resource definitions in a format compatible with Crossplane managed resources, similar in spirit to `kubectl get -o yaml`.
The goal is purely educational: by the end, you will have a working CLI skeleton and understand the core concepts of building exporters with `xp-clifford`.

This tutorial walks you through the project setup and the first steps step by step, building up the code incrementally. Later chapters of this tutorial will extend `kubexport` with real Kubernetes API calls.

---

## Table of Contents

1. [Prerequisites](#1-prerequisites)
2. [Creating a new Go project](#2-creating-a-new-go-project)
3. [Your first CLI](#3-your-first-cli)
4. [Implementing the export command](#4-implementing-the-export-command)
5. [Exporting a resource](#5-exporting-a-resource)
6. [Saving the output to a file](#6-saving-the-output-to-a-file)

---

## 1. Prerequisites

Before you begin, make sure you have the following installed:

- **Go 1.24 or later** — `xp-clifford` requires Go 1.24. Verify your version with:

  ```sh
  go version
  ```

  If you need to install or upgrade Go, follow the official instructions at <https://go.dev/doc/install>.

---

## 2. Creating a new Go project

Create a directory for your CLI tool and initialise a Go module inside it:

```sh
mkdir my-exporter
cd my-exporter
go mod init github.com/yourname/my-exporter
```

The module path (`github.com/yourname/my-exporter`) can be anything you like — it just needs to be a valid Go module path. If you are not planning to publish the module, a local path like `example.com/my-exporter` works equally well.

Now add `xp-clifford` as a dependency:

```sh
go get github.com/SAP/xp-clifford
```

This adds the module to your `go.mod`. Run `go mod tidy` to also populate `go.sum` with all required dependencies:

```sh
go mod tidy
```

Your project is now ready to use `xp-clifford`.

---

## 3. Your first CLI

Create a file called `main.go` in your project directory with the following content:

```go
package main

import (
    "github.com/SAP/xp-clifford/cli"
    _ "github.com/SAP/xp-clifford/cli/export"
)

func main() {
    cli.Configuration.ShortName = "kubexport"
    cli.Configuration.ObservedSystem = "Kubernetes"
    cli.Execute()
}
```

### What each line does

- `cli.Configuration.ShortName` — a short identifier for your tool, without spaces. It is used in the binary name (`kubexport-exporter`) and in the auto-generated configuration file name.
- `cli.Configuration.ObservedSystem` — the human-readable name of the system your tool exports resources from. It appears in the help text.
- `_ "github.com/SAP/xp-clifford/cli/export"` — this blank import registers the built-in `export` subcommand. The underscore means you only want the package's side effects (the registration), not to call anything from it directly.
- `cli.Execute()` — starts the CLI. Everything else is handled by the framework.

### Running it

```sh
go run main.go
```

You should see output similar to this:

```text
Kubernetes exporting tool is a CLI tool for exporting existing resources as Crossplane managed resources

Usage:
  kubexport-exporter [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  export      Export Kubernetes resources
  help        Help about any command

Flags:
  -c, --config string   Configuration file
  -h, --help            help for kubexport-exporter
  -v, --verbose         Verbose output

Use "kubexport-exporter [command] --help" for more information about a command.
```

The CLI is alive. It already has a help system, shell completion, and a global `--verbose` flag — all provided by the framework.

If you run the `export` subcommand now, you will get an error:

```sh
go run main.go export
```

```text
ERRO export subcommand is not set
```

That is expected. The `export` subcommand exists, but you have not told it what to do yet.

---

## 4. Implementing the export command

The `export` subcommand needs a function that contains your business logic. The function must have this exact signature:

```go
func(ctx context.Context, events export.EventHandler) error
```

- `ctx` — a standard Go context. Use it to detect cancellation (for example, when the user presses Ctrl-C).
- `events` — the communication channel back to the framework. You use it to report processed resources, warnings, and completion.
- The return value — return a non-nil error to signal a fatal failure.

Update `main.go` to add the export logic:

```go
package main

import (
    "context"
    "log/slog"

    "github.com/SAP/xp-clifford/cli"
    "github.com/SAP/xp-clifford/cli/export"
)

func exportLogic(_ context.Context, events export.EventHandler) error {
    slog.Info("export started")
    events.Stop()
    return nil
}

func main() {
    cli.Configuration.ShortName = "kubexport"
    cli.Configuration.ObservedSystem = "Kubernetes"
    export.SetCommand(exportLogic)
    cli.Execute()
}
```

Two things changed:

1. The `_ "github.com/SAP/xp-clifford/cli/export"` blank import was replaced by a regular import (`"github.com/SAP/xp-clifford/cli/export"`), because you now need to call `export.SetCommand` and use `export.EventHandler`.
2. `export.SetCommand(exportLogic)` registers your function with the framework before `cli.Execute()` is called.

Inside `exportLogic`, `events.Stop()` tells the framework that the export is finished. You must always call it — the framework will not clean up and exit properly without it.

Run the export subcommand:

```sh
go run main.go export
```

```text
INFO export started
```

The export command runs successfully. No resources are produced yet, but the structure is in place.

---

## 5. Exporting a resource

The `events.Resource(res)` method sends a resource to the framework for output. It accepts any value that implements the `resource.Object` interface from the Crossplane runtime — which is implemented by all Crossplane managed resource types.

For this tutorial, use `unstructured.Unstructured` from the Kubernetes API machinery library. It implements `resource.Object` and lets you build an arbitrary object from a plain Go map.

Update `main.go`:

```go
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
```

Run the export:

```sh
go run main.go export
```

```text
INFO export started


    ---
    password: secret
    user: test-user
    ...
```

The framework serialises the resource as YAML and prints it to standard output. The `---` and `...` markers are standard YAML document delimiters. If you call `events.Resource` multiple times, each resource is printed as a separate YAML document.

---

## 6. Saving the output to a file

The framework provides a built-in `-o` / `--output` flag on the `export` subcommand. Pass a file path to redirect the YAML output there instead of the terminal:

```sh
go run main.go export -o resources.yaml
```

```text
INFO export started
INFO Writing output to file output=resources.yaml
```

The log messages still appear on the terminal. Only the YAML resource output is written to the file:

```sh
cat resources.yaml
```

```text
---
password: secret
user: test-user
...
```

---

## Next steps

You have a working CLI tool that exports a resource. From here, you can:

- Replace the hardcoded `unstructured.Unstructured` with real API calls to the system you want to export from.
- Add configuration parameters (flags, environment variables, config file) using the `configparam` package — see the README for details.
- Add interactive prompts for credentials or selection using the `widget` package.
- Register additional subcommands (for example, a `login` command) using `cli.RegisterSubCommand`.

All of these are covered with full examples in the [README](../README.md).
