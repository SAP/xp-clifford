package configparam

import (
	"context"
	"fmt"

	"github.com/SAP/xp-clifford/cli/widget"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// FloatParam represents a configuration parameter that holds a float value.
type FloatParam struct {
	*configWithDefaultValue[FloatParam, float64]
}

// Ensure FloatParam implements the ConfigParam interface at compile time.
var _ ConfigParam = &FloatParam{}

// Float creates a new FloatParam with the specified name and description.
// The default value is set to 0.0.
func Float(name, description string) *FloatParam {
	p := &FloatParam{}
	p.configWithDefaultValue = newConfigWithDefaultValue(p, name, description, 0.0)
	return p
}

// AttachToCommand registers the float parameter as a persistent flag
// on the provided cobra command. This allows the parameter to be set
// via command-line arguments.
func (p *FloatParam) AttachToCommand(command *cobra.Command) {
	p.attachToCommand(command.PersistentFlags().Float64, command.PersistentFlags().Float64P)
}

// Value retrieves the current float value from the configuration.
func (p *FloatParam) Value() float64 {
	return p.value(viper.GetFloat64)
}

// ValueOrAsk returns the configured value if it has been set,
// otherwise it prompts the user for input interactively.
// Returns the float value and any error that occurred during input.
func (p *FloatParam) ValueOrAsk(ctx context.Context) (float64, error) {
	if p.configParam.IsSet() {
		return p.Value(), nil
	}
	return p.AskValue(ctx)
}

// inputPrompt generates the prompt string displayed to the user
// when requesting input for this parameter.
func (p *FloatParam) inputPrompt() string {
	return fmt.Sprintf("%s [%s]: ", p.configParam.Description, p.configParam.Name)
}

// askValue prompts the user for a float input using the widget system.
// It displays the parameter's description and example value as guidance.
func (p *FloatParam) askValue(ctx context.Context) (float64, error) {
	return widget.FloatInput(ctx,
		p.inputPrompt(),
		p.configParam.Example,
	)
}

// AskValue prompts the user for a float value.
// This method always prompts the user, regardless of whether a value
// has already been set. The entered value is persisted in the configuration.
// Returns the entered value and any error that occurred during input.
func (p *FloatParam) AskValue(ctx context.Context) (float64, error) {
	value, err := p.askValue(ctx)
	if err != nil {
		return 0, err
	}
	viper.Set(p.configParam.Name, value)
	return value, nil
}
