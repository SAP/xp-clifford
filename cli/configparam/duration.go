package configparam

import (
	"context"
	"fmt"
	"time"

	"github.com/SAP/xp-clifford/cli/widget"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// DurationParam represents a configuration parameter that holds a
// time.Duration value.
type DurationParam struct {
	*configWithDefaultValue[DurationParam, time.Duration]
}

// Ensure DurationParam implements the ConfigParam interface at
// compile time.
var _ ConfigParam = &DurationParam{}

// Duration creates a new DurationParam with the specified name and description.
// The default value is set to 0.
func Duration(name, description string) *DurationParam {
	p := &DurationParam{}
	p.configWithDefaultValue = newConfigWithDefaultValue(p, name, description, time.Duration(0))
	return p
}

// AttachToCommand registers the duration parameter as a persistent flag
// on the provided cobra command. This allows the parameter to be set
// via command-line arguments.
func (p *DurationParam) AttachToCommand(command *cobra.Command) {
	p.attachToCommand(command.PersistentFlags().Duration, command.PersistentFlags().DurationP)
}

// Value retrieves the current duration value from the configuration.
func (p *DurationParam) Value() time.Duration {
	return p.value(viper.GetDuration)
}

// ValueOrAsk returns the configured value if it has been set,
// otherwise it prompts the user for input interactively.
// Returns the duration value and any error that occurred during input.
func (p *DurationParam) ValueOrAsk(ctx context.Context) (time.Duration, error) {
	if p.configParam.IsSet() {
		return p.Value(), nil
	}
	return p.AskValue(ctx)
}

// inputPrompt generates the prompt string displayed to the user
// when requesting input for this parameter.
func (p *DurationParam) inputPrompt() string {
	return fmt.Sprintf("%s [%s]: ", p.configParam.Description, p.configParam.Name)
}

// askValue prompts the user for an duration input using the widget system.
// It displays the parameter's description and example value as guidance.
func (p *DurationParam) askValue(ctx context.Context) (time.Duration, error) {
	return widget.DurationInput(ctx,
		p.inputPrompt(),
		p.configParam.Example,
	)
}

// AskValue prompts the user for an duration value.
// This method always prompts the user, regardless of whether a value
// has already been set. The entered value is persisted in the configuration.
// Returns the entered value and any error that occurred during input.
func (p *DurationParam) AskValue(ctx context.Context) (time.Duration, error) {
	value, err := p.askValue(ctx)
	if err != nil {
		return 0, err
	}
	viper.Set(p.configParam.Name, value)
	return value, nil
}
