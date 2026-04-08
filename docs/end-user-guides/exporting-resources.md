# Exporting Resources

The `export` subcommand is mandatory but requires you to implement the export logic.

## Export Function Signature

```go
func(ctx context.Context, events export.EventHandler) error
```

- **`ctx`** — Use `ctx.Done()` to handle interrupts (e.g., Ctrl-C)
- **`events`** — Communicates progress to the framework via three methods:
  - `Warn(err error)` — Recoverable error, does not stop the export
  - `Resource(res resource.Object)` — A processed managed resource
  - `Stop()` — Signals completion; no further calls allowed
- **return** — Return a non-nil error to indicate a fatal failure

Register the function with `export.SetCommand`:

```go
export.SetCommand(exportLogic)
```

## Exporting a Single Resource

Use `events.Resource()` to output a resource. Any type implementing `resource.Object` works:

```go
func exportLogic(_ context.Context, events export.EventHandler) error {
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
```

Output is printed to stdout. Redirect to a file with `-o`:

```sh
test-exporter export -o output.yaml
```

## Displaying Warnings

Report non-fatal issues during export:

```go
events.Warn(errors.New("resource skipped due to missing field"))
```

Warnings appear on stderr but **not** in the output file.

## Commented Export

Problematic resources can be included in the output but commented out, preventing accidental application.

Wrap a resource with `yaml.NewResourceWithComment`:

```go
commentedResource := yaml.NewResourceWithComment(res)
commentedResource.SetComment("incomplete resource — do not apply")
events.Resource(commentedResource)
```

This produces:

```yaml
#
# incomplete resource — do not apply
#
# ---
# password: secret
# user: test-user
# ...
```

The wrapped resource must implement the `yaml.CommentedYAML` interface:

```go
type CommentedYAML interface {
    Comment() (string, bool)
}
```

`bool` indicates whether to comment out; `string` provides the comment message.
