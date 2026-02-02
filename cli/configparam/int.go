package configparam

import (
	"context"
	"fmt"

	"github.com/SAP/xp-clifford/cli/widget"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// IntParam represents a configuration parameter that holds an integer value.
// It embeds configWithDefaultValue to provide common configuration functionality
// with integer-specific behavior.
type IntParam struct {
	*configWithDefaultValue[IntParam, int]
}

// Ensure IntParam implements the ConfigParam interface at compile time.
var _ ConfigParam = &IntParam{}

// Int creates a new IntParam with the specified name and description.
// The default value is set to 0.
func Int(name, description string) *IntParam {
	p := &IntParam{}
	p.configWithDefaultValue = newConfigWithDefaultValue(p, name, description, 0)
	return p
}

// AttachToCommand registers the integer parameter as a persistent flag
// on the provided cobra command. This allows the parameter to be set
// via command-line arguments.
func (p *IntParam) AttachToCommand(command *cobra.Command) {
	p.attachToCommand(command.PersistentFlags().Int, command.PersistentFlags().IntP)
}

// Value retrieves the current integer value from the configuration.
func (p *IntParam) Value() int {
	return p.value(viper.GetInt)
}

// ValueOrAsk returns the configured value if it has been set,
// otherwise it prompts the user for input interactively.
// Returns the integer value and any error that occurred during input.
func (p *IntParam) ValueOrAsk(ctx context.Context) (int, error) {
	if p.configParam.IsSet() {
		return p.Value(), nil
	}
	return p.AskValue(ctx)
}

// inputPrompt generates the prompt string displayed to the user
// when requesting input for this parameter.
func (p *IntParam) inputPrompt() string {
	return fmt.Sprintf("%s [%s]: ", p.configParam.Description, p.configParam.Name)
}

// askValue prompts the user for an integer input using the widget system.
// It displays the parameter's description and example value as guidance.
func (p *IntParam) askValue(ctx context.Context) (int, error) {
	return widget.IntInput(ctx,
		p.inputPrompt(),
		p.configParam.Example,
	)
}

// AskValue prompts the user for an integer value and stores it in viper.
// This method always prompts the user, regardless of whether a value
// has already been set. The entered value is persisted in the configuration.
// Returns the entered value and any error that occurred during input.
func (p *IntParam) AskValue(ctx context.Context) (int, error) {
	value, err := p.askValue(ctx)
	if err != nil {
		return 0, err
	}
	viper.Set(p.configParam.Name, value)
	return value, nil
}
