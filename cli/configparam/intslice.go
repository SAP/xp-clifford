package configparam

import (
	"context"

	"github.com/SAP/xp-clifford/cli/widget"
	"github.com/SAP/xp-clifford/erratt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type intSlicePossibleValuesFnType func() ([]int, error)

// IntSliceParam type represents a configuration parameter that
// holds an []int value.
type IntSliceParam struct {
	*configWithDefaultValue[IntSliceParam, []int]
	possibleValues   []int
	possibleValuesFn intSlicePossibleValuesFnType
}

var _ ConfigParam = &IntSliceParam{}

func intSliceGenerator(name, description string) *IntSliceParam {
	p := &IntSliceParam{
		possibleValues: []int{},
	}
	p.configWithDefaultValue = newConfigWithDefaultValue(p, name, description, []int{})
	return p
}

// IntSlice creates an [IntSliceParam] value. The mandatory name
// and description parameters must be set.
func IntSlice(name, description string) *IntSliceParam {
	return intSliceGenerator(name, description)
}

// WithPossibleValues restricts the interactive selection to the
// supplied slice. When the CLI prompts the user, only these strings
// are offered as choices.
func (p *IntSliceParam) WithPossibleValues(values []int) *IntSliceParam {
	p.possibleValues = values
	return p
}

// WithPossibleValuesFn lazily supplies the valid choices for
// interactive selection. The given function is called when the CLI
// prompts the user. The returned strings are presented as options.
func (p *IntSliceParam) WithPossibleValuesFn(fn func() ([]int, error)) *IntSliceParam {
	p.possibleValuesFn = fn
	return p
}

// WithEnvVarName for IntSliceParam is not supported. See
// https://github.com/spf13/viper/issues/1611.
func (p *IntSliceParam) WithEnvVarName(_ string) *IntSliceParam {
	panic("IntSliceParam does not support the WithEnvVarName method")
}

// AttachToCommand registers the persistent int-slice flag (long form and
// optional short form) with the supplied [cobra.Command].
func (p *IntSliceParam) AttachToCommand(command *cobra.Command) {
	p.attachToCommand(command.PersistentFlags().IntSlice, command.PersistentFlags().IntSliceP)
}

// Value returns the user configured int slice. If the user has not
// configured any values, the default value is returned.
func (p *IntSliceParam) Value() []int {
	return p.value(viper.GetIntSlice)
}

// ValueOrAsk returns the configured slice or prompts the user to
// choose from the values supplied via [WithPossibleValues] or
// [WithPossibleValuesFn].  It fails if neither has been set.
//
// After successful selection the chosen slice is stored and the
// parameter is considered set.
func (p *IntSliceParam) ValueOrAsk(ctx context.Context) ([]int, error) {
	if p.IsSet() {
		return p.Value(), nil
	}
	if len(p.possibleValues) == 0 && p.possibleValuesFn == nil {
		return nil, erratt.New("StringSliceParam ValueOrAsk invoked but possibleValues are not set", "name", p.Name)
	}
	possibleValues := p.possibleValues
	if len(possibleValues) == 0 {
		var err error
		possibleValues, err = p.possibleValuesFn()
		if err != nil {
			return nil, erratt.Errorf("cannot get possible values: %w", err)
		}
		if len(possibleValues) == 0 {
			return []int{}, nil
		}
	}
	values, err := widget.MultiInputOfType(ctx,
		p.Description,
		possibleValues,
		widget.IntConverter,
	)
	if err != nil {
		return nil, err
	}
	viper.Set(p.Name, values)
	return values, nil
}
