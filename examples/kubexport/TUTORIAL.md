# Getting Started with xp-clifford

In this tutorial, you will build `kubexport` — a CLI tool that exports Kubernetes resource definitions in a format compatible with Crossplane managed resources, similar in spirit to `kubectl get -o yaml`.
The goal is purely educational: by the end, you will have a working CLI skeleton and understand the core concepts of building exporters with `xp-clifford`.

This tutorial walks you through the project setup and the first steps step by step, building up the code incrementally. Later chapters of this tutorial will extend `kubexport` with real Kubernetes API calls.

> **Note:** Each chapter shows the complete `main.go` at that step. The file committed in the repository reflects the final state of the tutorial. If you are following along, replace the entire file contents at each step.

---

## Table of Contents

1. [Prerequisites](#1-prerequisites)
2. [Creating a new Go project](#2-creating-a-new-go-project)
3. [Your first CLI](#3-your-first-cli)
4. [Implementing the export command](#4-implementing-the-export-command)
5. [Exporting a resource](#5-exporting-a-resource)
6. [Saving the output to a file](#6-saving-the-output-to-a-file)
7. [Connecting to Kubernetes using kubeconfig](#7-connecting-to-kubernetes-using-kubeconfig)
8. [Exporting real Namespace resources](#8-exporting-real-namespace-resources)
9. [Running the Namespace export](#9-running-the-namespace-export)
10. [Choosing what to export with the built-in kind parameter](#10-choosing-what-to-export-with-the-built-in-kind-parameter)
11. [Exporting ClusterRole resources](#11-exporting-clusterrole-resources)
12. [Exporting Pod resources](#12-exporting-pod-resources)
13. [Adding a namespace configuration parameter](#13-adding-a-namespace-configuration-parameter)
14. [Selecting a namespace interactively](#14-selecting-a-namespace-interactively)
15. [Collecting related resources with mkcontainer](#15-collecting-related-resources-with-mkcontainer)

---

## 1. Prerequisites

Before you begin, make sure you have the following installed:

- **Go 1.24 or later** — `xp-clifford` requires Go 1.24. Verify your version with:

  ```sh
  go version
  ```

  If you need to install or upgrade Go, follow the official instructions at <https://go.dev/doc/install>.

- **Access to a Kubernetes cluster from chapter 7 onward** — the later chapters talk to a live Kubernetes API through your kubeconfig. A local `kind` cluster works well for following along.

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

## 7. Connecting to Kubernetes using kubeconfig

The hardcoded object was useful to understand how `xp-clifford` works, but a real exporter needs to talk to an API, in this case Kubernetes.
For this tutorial, use the standard `client-go` kubeconfig loading rules. That means:

- If `$KUBECONFIG` is set, `client-go` uses it.
- Otherwise, it falls back to the default kubeconfig file, usually `~/.kube/config`.
- You do not need to add custom CLI flags for API connection settings.

In this chapter, you will replace the hardcoded `Unstructured` object with real Kubernetes client setup. The export logic itself will remain a placeholder — the first real API request comes in the next chapter.

Update `main.go`:

```go
package main

import (
    "context"
    "log/slog"

    "github.com/SAP/xp-clifford/cli"
    "github.com/SAP/xp-clifford/cli/export"
    "github.com/SAP/xp-clifford/erratt"

    "k8s.io/client-go/kubernetes"
    "k8s.io/client-go/tools/clientcmd"
)

func newClientset() (*kubernetes.Clientset, error) {
    loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
    clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, &clientcmd.ConfigOverrides{})

    restConfig, err := clientConfig.ClientConfig()
    if err != nil {
        return nil, erratt.Errorf("cannot load kubeconfig: %w", err)
    }

    clientset, err := kubernetes.NewForConfig(restConfig)
    if err != nil {
        return nil, erratt.Errorf("cannot create Kubernetes client: %w", err)
    }

    return clientset, nil
}

func exportLogic(ctx context.Context, events export.EventHandler) error {
    defer events.Stop()

    slog.Info("export started")

    clientset, err := newClientset()
    if err != nil {
        return err
    }

    slog.Info("prepared Kubernetes client", "host", clientset.RESTClient().Get().URL().Host)

    return nil
}

func main() {
    cli.Configuration.ShortName = "kubexport"
    cli.Configuration.ObservedSystem = "Kubernetes"
    export.SetCommand(exportLogic)
    cli.Execute()
}
```

Run `go mod tidy` once more so Go records the newly used Kubernetes client packages:

```sh
go mod tidy
```

### What changed in chapter 7

- `clientcmd.NewDefaultClientConfigLoadingRules()` loads Kubernetes API settings using the standard kubeconfig mechanism.
- `clientcmd.NewNonInteractiveDeferredLoadingClientConfig(...)` builds a client configuration without prompting the user.
- `kubernetes.NewForConfig(...)` creates a typed Kubernetes clientset.
- `clientset.RESTClient().Get().URL().Host` shows which API host the client is configured to target. It does not contact the cluster yet.
- `erratt.Errorf(...)` wraps errors with structured attributes — this is the recommended error handling pattern in `xp-clifford`.
- `defer events.Stop()` replaces the explicit `events.Stop()` call from previous chapters. It makes sure the framework is always notified when the export finishes, even if the function returns early because of an error.

The exporter does not produce any resources yet, but kubeconfig loading and client initialization are now wired up.

---

## 8. Exporting real Namespace resources

Now that the exporter can connect to the cluster, you will replace the placeholder log line with a real Kubernetes API call that lists and exports all `Namespace` objects.

Update `main.go`:

```go
package main

import (
    "context"
    "log/slog"

    "github.com/SAP/xp-clifford/cli"
    "github.com/SAP/xp-clifford/cli/export"
    "github.com/SAP/xp-clifford/erratt"

    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/client-go/kubernetes"
    "k8s.io/client-go/tools/clientcmd"
)

func newClientset() (*kubernetes.Clientset, error) {
    loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
    clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, &clientcmd.ConfigOverrides{})

    restConfig, err := clientConfig.ClientConfig()
    if err != nil {
        return nil, erratt.Errorf("cannot load kubeconfig: %w", err)
    }

    clientset, err := kubernetes.NewForConfig(restConfig)
    if err != nil {
        return nil, erratt.Errorf("cannot create Kubernetes client: %w", err)
    }

    return clientset, nil
}

func exportLogic(ctx context.Context, events export.EventHandler) error {
    defer events.Stop()

    slog.Info("export started")

    clientset, err := newClientset()
    if err != nil {
        return err
    }

    namespaces, err := clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
    if err != nil {
        return erratt.Errorf("cannot list namespaces: %w", err)
    }

    slog.Info("exporting namespaces", "count", len(namespaces.Items))
    for i := range namespaces.Items {
        namespace := namespaces.Items[i].DeepCopy()
        namespace.TypeMeta = metav1.TypeMeta{
            APIVersion: "v1",
            Kind:       "Namespace",
        }
        events.Resource(namespace)
    }

    return nil
}

func main() {
    cli.Configuration.ShortName = "kubexport"
    cli.Configuration.ObservedSystem = "Kubernetes"
    export.SetCommand(exportLogic)
    cli.Execute()
}
```

### What changed in chapter 8

- `clientset.CoreV1().Namespaces().List(...)` queries the Kubernetes API for all namespace objects.
- `namespace.DeepCopy()` creates a copy of each list item so the original response is not mutated.
- `namespace.TypeMeta = metav1.TypeMeta{...}` ensures the exported YAML contains `apiVersion` and `kind`, because Kubernetes list responses do not always populate those fields on each item.

---

## 9. Running the Namespace export

The exporter now talks to a real cluster and emits each namespace returned by the Kubernetes API.

Run the export command:

```sh
go run main.go export
```

If your kubeconfig points to a working cluster, you should see log output followed by YAML documents for the namespaces. The exact namespaces, counts, and metadata depend on your cluster:

```text
INFO export started
INFO exporting namespaces count=4

---
apiVersion: v1
kind: Namespace
metadata:
  creationTimestamp: "2026-04-02T09:12:34Z"
  name: default
  resourceVersion: "123"
  uid: 11111111-2222-3333-4444-555555555555
spec:
  finalizers:
  - kubernetes
status:
  phase: Active
...
```

The important part is that every object now comes from the live Kubernetes API.

The output flag still works exactly as before:

```sh
go run main.go export -o namespaces.yaml
```

This writes all exported Namespace objects to `namespaces.yaml` and keeps the log messages on the terminal.

---

## 10. Choosing what to export with the built-in kind parameter

So far, `kubexport` always exports namespaces.
That was fine while `Namespace` was the only supported resource kind, but the next feature will add more kinds.

The good news is that `xp-clifford` already provides a built-in `kind` configuration parameter on the `export` subcommand.
In this chapter, you will start using that parameter, even though the tool still supports only `Namespace`.

Update `main.go`:

```go
package main

import (
    "context"
    "log/slog"

    "github.com/SAP/xp-clifford/cli"
    "github.com/SAP/xp-clifford/cli/export"
    "github.com/SAP/xp-clifford/erratt"

    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/client-go/kubernetes"
    "k8s.io/client-go/tools/clientcmd"
)

func newClientset() (*kubernetes.Clientset, error) {
    loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
    clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, &clientcmd.ConfigOverrides{})

    restConfig, err := clientConfig.ClientConfig()
    if err != nil {
        return nil, erratt.Errorf("cannot load kubeconfig: %w", err)
    }

    clientset, err := kubernetes.NewForConfig(restConfig)
    if err != nil {
        return nil, erratt.Errorf("cannot create Kubernetes client: %w", err)
    }

    return clientset, nil
}

func exportNamespaces(ctx context.Context, clientset *kubernetes.Clientset, events export.EventHandler) error {
    namespaces, err := clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
    if err != nil {
        return erratt.Errorf("cannot list namespaces: %w", err)
    }

    slog.Info("exporting namespaces", "count", len(namespaces.Items))
    for i := range namespaces.Items {
        namespace := namespaces.Items[i].DeepCopy()
        namespace.TypeMeta = metav1.TypeMeta{
            APIVersion: "v1",
            Kind:       "Namespace",
        }
        events.Resource(namespace)
    }

    return nil
}

func exportLogic(ctx context.Context, events export.EventHandler) error {
    defer events.Stop()

    kinds, err := export.ResourceKindParam.ValueOrAsk(ctx)
    if err != nil {
        return erratt.Errorf("cannot determine resource kinds: %w", err)
    }
    if len(kinds) == 0 {
        return erratt.New("no resource kinds selected")
    }

    slog.Info("export started", "kind", kinds)

    clientset, err := newClientset()
    if err != nil {
        return err
    }

    for _, kind := range kinds {
        switch kind {
        case "Namespace":
            if err := exportNamespaces(ctx, clientset, events); err != nil {
                return err
            }
        default:
            return erratt.New("unsupported resource kind", "kind", kind)
        }
    }

    return nil
}

func main() {
    cli.Configuration.ShortName = "kubexport"
    cli.Configuration.ObservedSystem = "Kubernetes"
    export.AddResourceKinds("Namespace")
    export.SetCommand(exportLogic)
    cli.Execute()
}
```

### What changed in chapter 10

- `export.AddResourceKinds("Namespace")` registers the supported value for the built-in `--kind` parameter.
- `export.ResourceKindParam.ValueOrAsk(ctx)` reads the configured kinds, or prompts interactively if the user did not provide `--kind`.
- `exportLogic` no longer hardcodes one export path. Instead, it loops over the selected kinds and dispatches to the appropriate helper.
- `exportNamespaces(...)` keeps the namespace-specific Kubernetes API logic separate from the selection logic.
- The empty selection case is handled explicitly with `no resource kinds selected`.

Run the exporter and choose the resource kind explicitly:

```sh
go run main.go export --kind Namespace
```

You should see output similar to this:

```text
INFO export started kind=[Namespace]
INFO exporting namespaces count=4

---
apiVersion: v1
kind: Namespace
...
```

If you omit `--kind`, Clifford prompts interactively.
At this point the selector contains only one possible value: `Namespace`.

Even though there is only one supported kind so far, this is an important change: the exporter is now driven by user-selected kinds instead of one hardcoded export path.
The `kind` parameter is multi-valued, which is why the CLI uses repeated `--kind` flags and the interactive prompt is a multi-select list.

---

## 11. Exporting ClusterRole resources

The structure from the previous chapter is already flexible enough to support more resource kinds.
Now you will add a second exporter that reads `ClusterRole` objects from the Kubernetes RBAC API.

Update `main.go` again:

```go
package main

import (
    "context"
    "log/slog"

    "github.com/SAP/xp-clifford/cli"
    "github.com/SAP/xp-clifford/cli/export"
    "github.com/SAP/xp-clifford/erratt"

    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/client-go/kubernetes"
    "k8s.io/client-go/tools/clientcmd"
)

func newClientset() (*kubernetes.Clientset, error) {
    loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
    clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, &clientcmd.ConfigOverrides{})

    restConfig, err := clientConfig.ClientConfig()
    if err != nil {
        return nil, erratt.Errorf("cannot load kubeconfig: %w", err)
    }

    clientset, err := kubernetes.NewForConfig(restConfig)
    if err != nil {
        return nil, erratt.Errorf("cannot create Kubernetes client: %w", err)
    }

    return clientset, nil
}

func exportNamespaces(ctx context.Context, clientset *kubernetes.Clientset, events export.EventHandler) error {
    namespaces, err := clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
    if err != nil {
        return erratt.Errorf("cannot list namespaces: %w", err)
    }

    slog.Info("exporting namespaces", "count", len(namespaces.Items))
    for i := range namespaces.Items {
        namespace := namespaces.Items[i].DeepCopy()
        namespace.TypeMeta = metav1.TypeMeta{
            APIVersion: "v1",
            Kind:       "Namespace",
        }
        events.Resource(namespace)
    }

    return nil
}

func exportClusterRoles(ctx context.Context, clientset *kubernetes.Clientset, events export.EventHandler) error {
    clusterRoles, err := clientset.RbacV1().ClusterRoles().List(ctx, metav1.ListOptions{})
    if err != nil {
        return erratt.Errorf("cannot list cluster roles: %w", err)
    }

    slog.Info("exporting cluster roles", "count", len(clusterRoles.Items))
    for i := range clusterRoles.Items {
        clusterRole := clusterRoles.Items[i].DeepCopy()
        clusterRole.TypeMeta = metav1.TypeMeta{
            APIVersion: "rbac.authorization.k8s.io/v1",
            Kind:       "ClusterRole",
        }
        events.Resource(clusterRole)
    }

    return nil
}

func exportLogic(ctx context.Context, events export.EventHandler) error {
    defer events.Stop()

    kinds, err := export.ResourceKindParam.ValueOrAsk(ctx)
    if err != nil {
        return erratt.Errorf("cannot determine resource kinds: %w", err)
    }
    if len(kinds) == 0 {
        return erratt.New("no resource kinds selected")
    }

    slog.Info("export started", "kind", kinds)

    clientset, err := newClientset()
    if err != nil {
        return err
    }

    for _, kind := range kinds {
        switch kind {
        case "Namespace":
            if err := exportNamespaces(ctx, clientset, events); err != nil {
                return err
            }
        case "ClusterRole":
            if err := exportClusterRoles(ctx, clientset, events); err != nil {
                return err
            }
        default:
            return erratt.New("unsupported resource kind", "kind", kind)
        }
    }

    return nil
}

func main() {
    cli.Configuration.ShortName = "kubexport"
    cli.Configuration.ObservedSystem = "Kubernetes"
    export.AddResourceKinds("Namespace", "ClusterRole")
    export.SetCommand(exportLogic)
    cli.Execute()
}
```

### What changed in chapter 11

- `export.AddResourceKinds(...)` now registers both `Namespace` and `ClusterRole`.
- `exportClusterRoles(...)` adds a second exporter using `clientset.RbacV1().ClusterRoles().List(...)`.
- `ClusterRole` objects get `TypeMeta` before they are emitted, just like `Namespace` objects do.
- The `switch` in `exportLogic` now has two supported branches.

Export only cluster roles:

```sh
go run main.go export --kind ClusterRole
```

You should see output similar to this:

```text
INFO export started kind=[ClusterRole]
INFO exporting cluster roles count=72

---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
...
```

Export both supported kinds in one run:

```sh
go run main.go export --kind Namespace --kind ClusterRole -o resources.yaml
```

Because `kind` accepts multiple values, you pass both selections by repeating `--kind`.
The YAML output file now contains both `Namespace` and `ClusterRole` documents in the order you selected them.

Because both kinds are cluster-scoped, you still do not need a namespace parameter.
That becomes relevant in the next feature, when you add Pod exporting.

---

## 12. Exporting Pod resources

Until now, every supported resource kind has been cluster-scoped.
`Pod` is different: it is a namespaced resource, so the exporter must know which namespace to read from.

To keep this step focused on the new Kubernetes API call, start with a temporary hardcoded namespace.
Use `kube-system`, because it exists on all Kubernetes clusters and usually contains Pods on development clusters such as `kind`.

Update `main.go`:

```go
package main

import (
    "context"
    "log/slog"

    "github.com/SAP/xp-clifford/cli"
    "github.com/SAP/xp-clifford/cli/export"
    "github.com/SAP/xp-clifford/erratt"

    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/client-go/kubernetes"
    "k8s.io/client-go/tools/clientcmd"
)

func newClientset() (*kubernetes.Clientset, error) {
    loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
    clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, &clientcmd.ConfigOverrides{})

    restConfig, err := clientConfig.ClientConfig()
    if err != nil {
        return nil, erratt.Errorf("cannot load kubeconfig: %w", err)
    }

    clientset, err := kubernetes.NewForConfig(restConfig)
    if err != nil {
        return nil, erratt.Errorf("cannot create Kubernetes client: %w", err)
    }

    return clientset, nil
}

func exportNamespaces(ctx context.Context, clientset *kubernetes.Clientset, events export.EventHandler) error {
    namespaces, err := clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
    if err != nil {
        return erratt.Errorf("cannot list namespaces: %w", err)
    }

    slog.Info("exporting namespaces", "count", len(namespaces.Items))
    for i := range namespaces.Items {
        namespace := namespaces.Items[i].DeepCopy()
        namespace.TypeMeta = metav1.TypeMeta{
            APIVersion: "v1",
            Kind:       "Namespace",
        }
        events.Resource(namespace)
    }

    return nil
}

func exportClusterRoles(ctx context.Context, clientset *kubernetes.Clientset, events export.EventHandler) error {
    clusterRoles, err := clientset.RbacV1().ClusterRoles().List(ctx, metav1.ListOptions{})
    if err != nil {
        return erratt.Errorf("cannot list cluster roles: %w", err)
    }

    slog.Info("exporting cluster roles", "count", len(clusterRoles.Items))
    for i := range clusterRoles.Items {
        clusterRole := clusterRoles.Items[i].DeepCopy()
        clusterRole.TypeMeta = metav1.TypeMeta{
            APIVersion: "rbac.authorization.k8s.io/v1",
            Kind:       "ClusterRole",
        }
        events.Resource(clusterRole)
    }

    return nil
}

func exportPods(ctx context.Context, clientset *kubernetes.Clientset, namespace string, events export.EventHandler) error {
    pods, err := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
    if err != nil {
        return erratt.Errorf("cannot list pods: %w", err).With("namespace", namespace)
    }

    slog.Info("exporting pods", "namespace", namespace, "count", len(pods.Items))
    for i := range pods.Items {
        pod := pods.Items[i].DeepCopy()
        pod.TypeMeta = metav1.TypeMeta{
            APIVersion: "v1",
            Kind:       "Pod",
        }
        events.Resource(pod)
    }

    return nil
}

func exportLogic(ctx context.Context, events export.EventHandler) error {
    defer events.Stop()

    kinds, err := export.ResourceKindParam.ValueOrAsk(ctx)
    if err != nil {
        return erratt.Errorf("cannot determine resource kinds: %w", err)
    }
    if len(kinds) == 0 {
        return erratt.New("no resource kinds selected")
    }

    slog.Info("export started", "kind", kinds)

    clientset, err := newClientset()
    if err != nil {
        return err
    }

    for _, kind := range kinds {
        switch kind {
        case "Namespace":
            if err := exportNamespaces(ctx, clientset, events); err != nil {
                return err
            }
        case "ClusterRole":
            if err := exportClusterRoles(ctx, clientset, events); err != nil {
                return err
            }
        case "Pod":
            if err := exportPods(ctx, clientset, "kube-system", events); err != nil {
                return err
            }
        default:
            return erratt.New("unsupported resource kind", "kind", kind)
        }
    }

    return nil
}

func main() {
    cli.Configuration.ShortName = "kubexport"
    cli.Configuration.ObservedSystem = "Kubernetes"
    export.AddResourceKinds("Namespace", "ClusterRole", "Pod")
    export.SetCommand(exportLogic)
    cli.Execute()
}
```

### What changed in chapter 12

- `export.AddResourceKinds(...)` now registers `Pod` in addition to the cluster-scoped kinds.
- `exportPods(...)` adds a third exporter using `clientset.CoreV1().Pods(namespace).List(...)`.
- The exporter sets `TypeMeta` on each Pod so the YAML includes `apiVersion: v1` and `kind: Pod`.
- For this first Pod-enabled version, the namespace is hardcoded to `kube-system`.

Run the exporter:

```sh
go run main.go export --kind Pod
```

You should see output similar to this. The exact pod count and metadata depend on your cluster:

```text
INFO export started kind=[Pod]
INFO exporting pods namespace=kube-system count=4

---
apiVersion: v1
kind: Pod
...
```

This proves the exporter can now handle a namespaced resource kind.
The obvious limitation is that the namespace is still fixed in the source code.

---

## 13. Adding a namespace configuration parameter

Hardcoding `kube-system` was useful for the first Pod example, but a real exporter needs the namespace to be configurable.

In this chapter, add a real `namespace` configuration parameter.
The user can set it with a flag, environment variable, or config file, just like any other Clifford configuration parameter.

Update `main.go`:

```go
package main

import (
    "context"
    "log/slog"

    "github.com/SAP/xp-clifford/cli"
    "github.com/SAP/xp-clifford/cli/configparam"
    "github.com/SAP/xp-clifford/cli/export"
    "github.com/SAP/xp-clifford/erratt"

    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/client-go/kubernetes"
    "k8s.io/client-go/tools/clientcmd"
)

func newClientset() (*kubernetes.Clientset, error) {
    loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
    clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, &clientcmd.ConfigOverrides{})

    restConfig, err := clientConfig.ClientConfig()
    if err != nil {
        return nil, erratt.Errorf("cannot load kubeconfig: %w", err)
    }

    clientset, err := kubernetes.NewForConfig(restConfig)
    if err != nil {
        return nil, erratt.Errorf("cannot create Kubernetes client: %w", err)
    }

    return clientset, nil
}

func exportNamespaces(ctx context.Context, clientset *kubernetes.Clientset, events export.EventHandler) error {
    namespaces, err := clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
    if err != nil {
        return erratt.Errorf("cannot list namespaces: %w", err)
    }

    slog.Info("exporting namespaces", "count", len(namespaces.Items))
    for i := range namespaces.Items {
        namespace := namespaces.Items[i].DeepCopy()
        namespace.TypeMeta = metav1.TypeMeta{
            APIVersion: "v1",
            Kind:       "Namespace",
        }
        events.Resource(namespace)
    }

    return nil
}

func exportClusterRoles(ctx context.Context, clientset *kubernetes.Clientset, events export.EventHandler) error {
    clusterRoles, err := clientset.RbacV1().ClusterRoles().List(ctx, metav1.ListOptions{})
    if err != nil {
        return erratt.Errorf("cannot list cluster roles: %w", err)
    }

    slog.Info("exporting cluster roles", "count", len(clusterRoles.Items))
    for i := range clusterRoles.Items {
        clusterRole := clusterRoles.Items[i].DeepCopy()
        clusterRole.TypeMeta = metav1.TypeMeta{
            APIVersion: "rbac.authorization.k8s.io/v1",
            Kind:       "ClusterRole",
        }
        events.Resource(clusterRole)
    }

    return nil
}

func exportPods(ctx context.Context, clientset *kubernetes.Clientset, namespace string, events export.EventHandler) error {
    pods, err := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
    if err != nil {
        return erratt.Errorf("cannot list pods: %w", err).With("namespace", namespace)
    }

    slog.Info("exporting pods", "namespace", namespace, "count", len(pods.Items))
    for i := range pods.Items {
        pod := pods.Items[i].DeepCopy()
        pod.TypeMeta = metav1.TypeMeta{
            APIVersion: "v1",
            Kind:       "Pod",
        }
        events.Resource(pod)
    }

    return nil
}

func exportLogic(ctx context.Context, events export.EventHandler) error {
    defer events.Stop()

    kinds, err := export.ResourceKindParam.ValueOrAsk(ctx)
    if err != nil {
        return erratt.Errorf("cannot determine resource kinds: %w", err)
    }
    if len(kinds) == 0 {
        return erratt.New("no resource kinds selected")
    }

    podNamespace := ""
    for _, kind := range kinds {
        if kind == "Pod" {
            podNamespace = namespaceParam.Value()
            if podNamespace == "" {
                return erratt.New("namespace must be configured for Pod export")
            }
            break
        }
    }

    if podNamespace != "" {
        slog.Info("export started", "kind", kinds, "namespace", podNamespace)
    } else {
        slog.Info("export started", "kind", kinds)
    }

    clientset, err := newClientset()
    if err != nil {
        return err
    }

    for _, kind := range kinds {
        switch kind {
        case "Namespace":
            if err := exportNamespaces(ctx, clientset, events); err != nil {
                return err
            }
        case "ClusterRole":
            if err := exportClusterRoles(ctx, clientset, events); err != nil {
                return err
            }
        case "Pod":
            if err := exportPods(ctx, clientset, podNamespace, events); err != nil {
                return err
            }
        default:
            return erratt.New("unsupported resource kind", "kind", kind)
        }
    }

    return nil
}

var namespaceParam = configparam.String("namespace", "Namespace for namespaced resources such as Pod").
    WithShortName("n").
    WithEnvVarName("NAMESPACE")

func main() {
    cli.Configuration.ShortName = "kubexport"
    cli.Configuration.ObservedSystem = "Kubernetes"
    export.AddConfigParams(namespaceParam)
    export.AddResourceKinds("Namespace", "ClusterRole", "Pod")
    export.SetCommand(exportLogic)
    cli.Execute()
}
```

### What changed in chapter 13

- `namespaceParam` defines a real string configuration parameter for Pod export.
- `export.AddConfigParams(namespaceParam)` registers the parameter on the built-in `export` subcommand.
- The Pod branch now reads the namespace from configuration instead of using a hardcoded value.
- The exporter fails with `namespace must be configured for Pod export` if `Pod` is selected but no namespace is configured yet.

Export Pods by passing the namespace explicitly:

```sh
go run main.go export --kind Pod --namespace kube-system
```

You can also supply the same value using `NAMESPACE=kube-system` or a config file, because `namespaceParam` is a standard Clifford configuration parameter.

This is already a useful version of the tool, but it still has one limitation: when the namespace is not configured, the user gets an error instead of a guided selection.

---

## 14. Selecting a namespace interactively

The final step is to remove that limitation.
When `Pod` export is requested and no namespace is configured, the exporter should query the cluster for namespaces and let the user choose one interactively.

Because `configparam.String` only provides a text prompt, use the lower-level `widget.MultiInput(...)` helper directly here.
It is still a multi-select widget, so the code must enforce that exactly one namespace is chosen.

Update `main.go` one last time:

```go
package main

import (
    "context"
    "log/slog"
    "sort"

    "github.com/SAP/xp-clifford/cli"
    "github.com/SAP/xp-clifford/cli/configparam"
    "github.com/SAP/xp-clifford/cli/export"
    "github.com/SAP/xp-clifford/cli/widget"
    "github.com/SAP/xp-clifford/erratt"

    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/client-go/kubernetes"
    "k8s.io/client-go/tools/clientcmd"
)

func newClientset() (*kubernetes.Clientset, error) {
    loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
    clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, &clientcmd.ConfigOverrides{})

    restConfig, err := clientConfig.ClientConfig()
    if err != nil {
        return nil, erratt.Errorf("cannot load kubeconfig: %w", err)
    }

    clientset, err := kubernetes.NewForConfig(restConfig)
    if err != nil {
        return nil, erratt.Errorf("cannot create Kubernetes client: %w", err)
    }

    return clientset, nil
}

func uniqueStrings(values []string) []string {
    seen := make(map[string]struct{}, len(values))
    unique := make([]string, 0, len(values))
    for _, value := range values {
        if _, ok := seen[value]; ok {
            continue
        }
        seen[value] = struct{}{}
        unique = append(unique, value)
    }
    return unique
}

func listNamespaceNames(ctx context.Context, clientset *kubernetes.Clientset) ([]string, error) {
    namespaces, err := clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
    if err != nil {
        return nil, erratt.Errorf("cannot list namespaces for selection: %w", err)
    }

    namespaceNames := make([]string, len(namespaces.Items))
    for i := range namespaces.Items {
        namespaceNames[i] = namespaces.Items[i].GetName()
    }
    sort.Strings(namespaceNames)
    return namespaceNames, nil
}

func resolveNamespace(ctx context.Context, clientset *kubernetes.Clientset) (string, error) {
    if namespace := namespaceParam.Value(); namespace != "" {
        return namespace, nil
    }

    namespaceNames, err := listNamespaceNames(ctx, clientset)
    if err != nil {
        return "", err
    }
    if len(namespaceNames) == 0 {
        return "", erratt.New("cannot select namespace: no namespaces available")
    }

    selected, err := widget.MultiInput(ctx, "Select namespace for Pod export", namespaceNames)
    if err != nil {
        return "", erratt.Errorf("cannot select namespace: %w", err)
    }
    if len(selected) != 1 {
        return "", erratt.New("select exactly one namespace for Pod export", "selected", selected)
    }

    return selected[0], nil
}

func exportNamespaces(ctx context.Context, clientset *kubernetes.Clientset, events export.EventHandler) error {
    namespaces, err := clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
    if err != nil {
        return erratt.Errorf("cannot list namespaces: %w", err)
    }

    slog.Info("exporting namespaces", "count", len(namespaces.Items))
    for i := range namespaces.Items {
        namespace := namespaces.Items[i].DeepCopy()
        namespace.TypeMeta = metav1.TypeMeta{
            APIVersion: "v1",
            Kind:       "Namespace",
        }
        events.Resource(namespace)
    }

    return nil
}

func exportClusterRoles(ctx context.Context, clientset *kubernetes.Clientset, events export.EventHandler) error {
    clusterRoles, err := clientset.RbacV1().ClusterRoles().List(ctx, metav1.ListOptions{})
    if err != nil {
        return erratt.Errorf("cannot list cluster roles: %w", err)
    }

    slog.Info("exporting cluster roles", "count", len(clusterRoles.Items))
    for i := range clusterRoles.Items {
        clusterRole := clusterRoles.Items[i].DeepCopy()
        clusterRole.TypeMeta = metav1.TypeMeta{
            APIVersion: "rbac.authorization.k8s.io/v1",
            Kind:       "ClusterRole",
        }
        events.Resource(clusterRole)
    }

    return nil
}

func exportPods(ctx context.Context, clientset *kubernetes.Clientset, namespace string, events export.EventHandler) error {
    pods, err := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
    if err != nil {
        return erratt.Errorf("cannot list pods: %w", err).With("namespace", namespace)
    }

    slog.Info("exporting pods", "namespace", namespace, "count", len(pods.Items))
    for i := range pods.Items {
        pod := pods.Items[i].DeepCopy()
        pod.TypeMeta = metav1.TypeMeta{
            APIVersion: "v1",
            Kind:       "Pod",
        }
        events.Resource(pod)
    }

    return nil
}

func exportLogic(ctx context.Context, events export.EventHandler) error {
    defer events.Stop()

    kinds, err := export.ResourceKindParam.ValueOrAsk(ctx)
    if err != nil {
        return erratt.Errorf("cannot determine resource kinds: %w", err)
    }
    kinds = uniqueStrings(kinds)
    if len(kinds) == 0 {
        return erratt.New("no resource kinds selected")
    }

    clientset, err := newClientset()
    if err != nil {
        return err
    }

    podNamespace := ""
    for _, kind := range kinds {
        if kind == "Pod" {
            podNamespace, err = resolveNamespace(ctx, clientset)
            if err != nil {
                return err
            }
            break
        }
    }

    if podNamespace != "" {
        slog.Info("export started", "kind", kinds, "namespace", podNamespace)
    } else {
        slog.Info("export started", "kind", kinds)
    }

    for _, kind := range kinds {
        switch kind {
        case "Namespace":
            if err := exportNamespaces(ctx, clientset, events); err != nil {
                return err
            }
        case "ClusterRole":
            if err := exportClusterRoles(ctx, clientset, events); err != nil {
                return err
            }
        case "Pod":
            if err := exportPods(ctx, clientset, podNamespace, events); err != nil {
                return err
            }
        default:
            return erratt.New("unsupported resource kind", "kind", kind)
        }
    }

    return nil
}

var namespaceParam = configparam.String("namespace", "Namespace for namespaced resources such as Pod").
    WithShortName("n").
    WithEnvVarName("NAMESPACE")

func main() {
    cli.Configuration.ShortName = "kubexport"
    cli.Configuration.ObservedSystem = "Kubernetes"
    export.AddConfigParams(namespaceParam)
    export.AddResourceKinds("Namespace", "ClusterRole", "Pod")
    export.SetCommand(exportLogic)
    cli.Execute()
}
```

### What changed in chapter 14

- `listNamespaceNames(...)` reads the available namespaces from the Kubernetes API and sorts them for a stable prompt order.
- `resolveNamespace(...)` uses the configured namespace if present, otherwise it prompts interactively from the live namespace list.
- `widget.MultiInput(...)` is used directly because `configparam.String` only provides text input, while this tutorial wants a selection from discovered values.
- `uniqueStrings(...)` removes duplicate kind selections while preserving their order.
- `exportLogic` resolves the namespace only when `Pod` export is requested.

The explicit configuration path still works:

```sh
go run main.go export --kind Pod --namespace kube-system
```

If you omit `--namespace`, the exporter queries the cluster and opens an interactive selector:

```sh
go run main.go export --kind Pod
```

Choose exactly one namespace from the list.
After the selection, the exporter continues with output similar to this. The exact pod count depends on your cluster:

```text
INFO export started kind=[Pod] namespace=kube-system
INFO exporting pods namespace=kube-system count=4

---
apiVersion: v1
kind: Pod
...
```

You can also combine `Pod` with the previously implemented kinds:

```sh
go run main.go export --kind Namespace --kind Pod --namespace kube-system -o resources.yaml
```

Because the exporter processes kinds in the order you select them, the YAML file contains all `Namespace` documents first, followed by the `Pod` documents.

---

## 15. Collecting related resources with mkcontainer

The previous chapter still had one artificial limitation: `Pod` export accepted exactly one namespace.
That kept the example small, but it was not a meaningful restriction.

In this final chapter, you will remove that limitation and add a practical use for `mkcontainer`.
The exporter will now:

- accept one or more `namespace` values for `Pod` export
- allow selecting multiple namespaces interactively
- optionally include the related `Namespace` resources with `--include-namespaces`
- deduplicate collected resources before emitting them

This is a good fit for `mkcontainer`.
Instead of emitting resources immediately, the exporter first collects them into an inventory.
That inventory uses:

- an ordered slice to preserve output order
- `mkcontainer.TypedContainer[...]` to index collected resources by Kubernetes UID and a logical key

The `namespace` parameter also becomes a `StringSlice` configuration parameter in this chapter.
That means all configuration paths stay consistent:

- repeat `--namespace` on the command line
- set `NAMESPACES` in the environment
- use a YAML list in the config file
- select multiple namespaces interactively when no value is configured

Update `main.go` one last time:

```go
package main

import (
    "context"
    "log/slog"
    "sort"

    "github.com/SAP/xp-clifford/cli"
    "github.com/SAP/xp-clifford/cli/configparam"
    "github.com/SAP/xp-clifford/cli/export"
    "github.com/SAP/xp-clifford/erratt"
    "github.com/SAP/xp-clifford/mkcontainer"

    "github.com/crossplane/crossplane-runtime/pkg/resource"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/client-go/kubernetes"
    "k8s.io/client-go/tools/clientcmd"
)

type collectedResource struct {
    key    string
    object resource.Object
}

func (r *collectedResource) GetGUID() string {
    return string(r.object.GetUID())
}

func (r *collectedResource) GetName() string {
    return r.key
}

type resourceInventory struct {
    ordered []*collectedResource
    index   mkcontainer.TypedContainer[*collectedResource]
}

func newResourceInventory() *resourceInventory {
    return &resourceInventory{
        ordered: make([]*collectedResource, 0),
        index:   mkcontainer.NewTyped[*collectedResource](),
    }
}

func resourceKey(obj resource.Object) string {
    if namespace := obj.GetNamespace(); namespace != "" {
        return obj.GetObjectKind().GroupVersionKind().Kind + "/" + namespace + "/" + obj.GetName()
    }
    return obj.GetObjectKind().GroupVersionKind().Kind + "/" + obj.GetName()
}

func namespaceKey(namespace string) string {
    return "Namespace/" + namespace
}

func (i *resourceInventory) HasKey(key string) bool {
    return len(i.index.GetByName(key)) > 0
}

func (i *resourceInventory) Add(obj resource.Object) bool {
    item := &collectedResource{
        key:    resourceKey(obj),
        object: obj,
    }
    if guid := item.GetGUID(); guid != "" && i.index.GetByGUID(guid) != nil {
        return false
    }
    if i.HasKey(item.GetName()) {
        return false
    }

    i.index.Store(item)
    i.ordered = append(i.ordered, item)
    return true
}

func (i *resourceInventory) Emit(events export.EventHandler) {
    for _, item := range i.ordered {
        events.Resource(item.object)
    }
}

func newClientset() (*kubernetes.Clientset, error) {
    loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
    clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, &clientcmd.ConfigOverrides{})

    restConfig, err := clientConfig.ClientConfig()
    if err != nil {
        return nil, erratt.Errorf("cannot load kubeconfig: %w", err)
    }

    clientset, err := kubernetes.NewForConfig(restConfig)
    if err != nil {
        return nil, erratt.Errorf("cannot create Kubernetes client: %w", err)
    }

    return clientset, nil
}

func uniqueStrings(values []string) []string {
    seen := make(map[string]struct{}, len(values))
    unique := make([]string, 0, len(values))
    for _, value := range values {
        if _, ok := seen[value]; ok {
            continue
        }
        seen[value] = struct{}{}
        unique = append(unique, value)
    }
    return unique
}

func listNamespaceNames(ctx context.Context, clientset *kubernetes.Clientset) ([]string, error) {
    namespaces, err := clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
    if err != nil {
        return nil, erratt.Errorf("cannot list namespaces for selection: %w", err)
    }

    namespaceNames := make([]string, len(namespaces.Items))
    for i := range namespaces.Items {
        namespaceNames[i] = namespaces.Items[i].GetName()
    }
    sort.Strings(namespaceNames)
    return namespaceNames, nil
}

func resolveNamespaces(ctx context.Context, clientset *kubernetes.Clientset) ([]string, error) {
    if namespaces := uniqueStrings(namespaceParam.Value()); len(namespaces) > 0 {
        return namespaces, nil
    }

    namespaceNames, err := listNamespaceNames(ctx, clientset)
    if err != nil {
        return nil, err
    }
    if len(namespaceNames) == 0 {
        return nil, erratt.New("cannot select namespaces: no namespaces available")
    }

    namespaceParam.WithPossibleValues(namespaceNames)
    namespaces, err := namespaceParam.ValueOrAsk(ctx)
    if err != nil {
        return nil, erratt.Errorf("cannot determine namespaces: %w", err)
    }
    namespaces = uniqueStrings(namespaces)
    if len(namespaces) == 0 {
        return nil, erratt.New("no namespaces selected for Pod export")
    }

    return namespaces, nil
}

func collectSelectedNamespace(ctx context.Context, clientset *kubernetes.Clientset, namespace string, inventory *resourceInventory) error {
    if inventory.HasKey(namespaceKey(namespace)) {
        return nil
    }

    namespaceResource, err := clientset.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
    if err != nil {
        return erratt.Errorf("cannot get namespace: %w", err).With("namespace", namespace)
    }

    namespaceResource.TypeMeta = metav1.TypeMeta{
        APIVersion: "v1",
        Kind:       "Namespace",
    }
    slog.Info("exporting selected namespace", "namespace", namespace)
    inventory.Add(namespaceResource)
    return nil
}

func collectNamespaces(ctx context.Context, clientset *kubernetes.Clientset, inventory *resourceInventory) error {
    namespaces, err := clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
    if err != nil {
        return erratt.Errorf("cannot list namespaces: %w", err)
    }

    slog.Info("exporting namespaces", "count", len(namespaces.Items))
    for i := range namespaces.Items {
        namespace := namespaces.Items[i].DeepCopy()
        namespace.TypeMeta = metav1.TypeMeta{
            APIVersion: "v1",
            Kind:       "Namespace",
        }
        inventory.Add(namespace)
    }

    return nil
}

func collectClusterRoles(ctx context.Context, clientset *kubernetes.Clientset, inventory *resourceInventory) error {
    clusterRoles, err := clientset.RbacV1().ClusterRoles().List(ctx, metav1.ListOptions{})
    if err != nil {
        return erratt.Errorf("cannot list cluster roles: %w", err)
    }

    slog.Info("exporting cluster roles", "count", len(clusterRoles.Items))
    for i := range clusterRoles.Items {
        clusterRole := clusterRoles.Items[i].DeepCopy()
        clusterRole.TypeMeta = metav1.TypeMeta{
            APIVersion: "rbac.authorization.k8s.io/v1",
            Kind:       "ClusterRole",
        }
        inventory.Add(clusterRole)
    }

    return nil
}

func collectPods(ctx context.Context, clientset *kubernetes.Clientset, namespace string, inventory *resourceInventory) error {
    pods, err := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
    if err != nil {
        return erratt.Errorf("cannot list pods: %w", err).With("namespace", namespace)
    }

    slog.Info("exporting pods", "namespace", namespace, "count", len(pods.Items))
    for i := range pods.Items {
        pod := pods.Items[i].DeepCopy()
        pod.TypeMeta = metav1.TypeMeta{
            APIVersion: "v1",
            Kind:       "Pod",
        }
        inventory.Add(pod)
    }

    return nil
}

func exportLogic(ctx context.Context, events export.EventHandler) error {
    defer events.Stop()

    kinds, err := export.ResourceKindParam.ValueOrAsk(ctx)
    if err != nil {
        return erratt.Errorf("cannot determine resource kinds: %w", err)
    }
    kinds = uniqueStrings(kinds)
    if len(kinds) == 0 {
        return erratt.New("no resource kinds selected")
    }

    clientset, err := newClientset()
    if err != nil {
        return err
    }

    podNamespaces := []string{}
    for _, kind := range kinds {
        if kind == "Pod" {
            podNamespaces, err = resolveNamespaces(ctx, clientset)
            if err != nil {
                return err
            }
            break
        }
    }

    if len(podNamespaces) > 0 {
        slog.Info("export started", "kind", kinds, "namespaces", podNamespaces)
    } else {
        slog.Info("export started", "kind", kinds)
    }

    inventory := newResourceInventory()

    for _, kind := range kinds {
        switch kind {
        case "Namespace":
            if err := collectNamespaces(ctx, clientset, inventory); err != nil {
                return err
            }
        case "ClusterRole":
            if err := collectClusterRoles(ctx, clientset, inventory); err != nil {
                return err
            }
        case "Pod":
            for _, namespace := range podNamespaces {
                if includeNamespacesParam.Value() {
                    if err := collectSelectedNamespace(ctx, clientset, namespace, inventory); err != nil {
                        return err
                    }
                }
                if err := collectPods(ctx, clientset, namespace, inventory); err != nil {
                    return err
                }
            }
        default:
            return erratt.New("unsupported resource kind", "kind", kind)
        }
    }
    inventory.Emit(events)

    return nil
}

var namespaceParam = configparam.StringSlice("namespace", "Namespaces for namespaced resources such as Pod").
    WithShortName("n").
    WithEnvVarName("NAMESPACES")

var includeNamespacesParam = configparam.Bool("include-namespaces", "Also export selected Namespace resources when exporting Pod")

func main() {
    cli.Configuration.ShortName = "kubexport"
    cli.Configuration.ObservedSystem = "Kubernetes"
    export.AddConfigParams(namespaceParam, includeNamespacesParam)
    export.AddResourceKinds("Namespace", "ClusterRole", "Pod")
    export.SetCommand(exportLogic)
    cli.Execute()
}
```

### What changed in chapter 15

- `namespaceParam` is now a `StringSlice` configuration parameter, so `Pod` export can work with one or more namespaces.
- `resolveNamespaces(...)` uses `StringSliceParam.ValueOrAsk(...)` together with the live namespace list, so flags, environment variables, config files, and interactive selection all support multiple namespaces consistently.
- `collectedResource` wraps a Kubernetes object so it can be stored in `mkcontainer` using both UID (`GetGUID`) and a logical key (`GetName`).
- `resourceInventory` combines an ordered slice with `mkcontainer.TypedContainer[...]`: the slice preserves output order, while `mkcontainer` prevents duplicate collection.
- `collectSelectedNamespace(...)` adds related Namespace objects only when `--include-namespaces` is enabled and skips them if they are already in the inventory.
- The resource-specific helpers now gather objects into the inventory instead of emitting them immediately, and `inventory.Emit(...)` performs the final output pass.

Export Pods from more than one namespace by repeating `--namespace`:

```sh
go run main.go export --kind Pod --namespace kube-system --namespace default
```

The exporter processes the selected namespaces in the order you provided.
On many clusters, some namespaces may contain no Pods, which is fine.

You can configure the same selection through the environment as well:

```sh
NAMESPACES="kube-system default" go run main.go export --kind Pod
```

The interactive prompt now also allows selecting multiple namespaces, because `namespaceParam` itself is multi-valued.

Run the related-resource mode with a single namespace:

```sh
go run main.go export --kind Pod --namespace kube-system --include-namespaces
```

You should see output similar to this. The exact pod count depends on your cluster:

```text
INFO export started kind=[Pod] namespaces=[kube-system]
INFO exporting selected namespace namespace=kube-system
INFO exporting pods namespace=kube-system count=4

---
apiVersion: v1
kind: Namespace
...
---
apiVersion: v1
kind: Pod
...
```

You can also combine the new flag with the explicit `Namespace` exporter:

```sh
go run main.go export --kind Namespace --kind Pod --namespace kube-system --namespace default --include-namespaces -o resources.yaml
```

In that case, the exporter still emits each selected namespace only once, even when a namespace is reached from two paths:

- the explicit `Namespace` export
- the Pod-related namespace inclusion enabled by `--include-namespaces`

Because the inventory preserves first-seen order, resources are emitted in the order they were first collected.
This makes the output deterministic while still allowing multiple collectors to overlap safely.

---

## Next steps

You now have a working CLI tool that can export cluster-scoped and namespaced Kubernetes resources, resolve one or more namespaces interactively, and collect related resources without duplicating them. From here, you can:

- Add more namespaced resource kinds by reusing the `namespace` resolution pattern from `Pod` export.
- Auto-include other related resources such as `ServiceAccount`, `ConfigMap`, or `Secret` by extending the inventory pattern from `mkcontainer`.
- Add filters such as labels or field selectors to reduce the exported result set.
- Add separate selection rules for different resources if some kinds should allow only one namespace while others should allow many.
- Register additional subcommands (for example, a `login` command) using `cli.RegisterSubCommand`.

See other examples in the [README](../../README.md).
