# Interactive Widgets

`xp-clifford` provides terminal UI widgets for user interaction. Widgets require an interactive terminal — when running in an IDE, enable terminal emulation (e.g., GoLand: *Emulate terminal in output console*).

## TextInput

Prompts for a single line of text.

```go
func TextInput(ctx context.Context, title, placeholder string, sensitive bool) (string, error)
```

- **`title`** — Prompt displayed to the user
- **`placeholder`** — Placeholder text when input is empty
- **`sensitive`** — When `true`, masks typed characters (e.g., for passwords)

```go
username, err := widget.TextInput(ctx, "Username", "anonymous", false)
password, err := widget.TextInput(ctx, "Password", "", true)
```

## IntInput / FloatInput / DurationInput

Work like `TextInput` but accept only `int`, `float64`, and `time.Duration` values respectively. They do not support the `sensitive` parameter.

## MultiInput

Multi-selection interface for choosing items from a predefined list.

```go
func MultiInput(ctx context.Context, title string, options []string) ([]string, error)
```

```go
protocols, err := widget.MultiInput(ctx,
    "Select the supported protocols",
    []string{"FTP", "HTTP", "HTTPS", "SFTP", "SSH"},
)
```

## Integration with Configuration Parameters

Widgets integrate with configuration parameters via `ValueOrAsk(ctx)`:

- `StringParam.ValueOrAsk(ctx)` — uses `TextInput`
- `StringSliceParam.ValueOrAsk(ctx)` — uses `MultiInput` (requires `WithPossibleValues` or `WithPossibleValuesFn`)

See [Configuration](./configuration.md) for details.
