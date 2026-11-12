package configparam

import (
	"context"

	"github.com/SAP/xp-clifford/cli/widget"
	"github.com/SAP/xp-clifford/erratt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type intSlicePossibleValuesFnType func() ([]int, error)
type intSliceTreeOptionsFnType func() ([]widget.TreeOptionOfType[int], error)

// IntSliceParam type represents a configuration parameter that
// holds an []int value.
type IntSliceParam struct {
	*configWithDefaultValue[IntSliceParam, []int]
	possibleValues   []int
	possibleValuesFn intSlicePossibleValuesFnType
	treeOptionsFn    intSliceTreeOptionsFnType
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
// are offered as choices. This switches to flat selection.
func (p *IntSliceParam) WithPossibleValues(values []int) *IntSliceParam {
	p.possibleValues = values
	p.treeOptionsFn = nil
	return p
}

// WithPossibleValuesFn lazily supplies the valid choices for
// interactive selection. The given function is called when the CLI
// prompts the user. The returned strings are presented as options.
// This switches to flat selection.
func (p *IntSliceParam) WithPossibleValuesFn(fn func() ([]int, error)) *IntSliceParam {
	p.possibleValuesFn = fn
	p.treeOptionsFn = nil
	return p
}

// WithTreeOptions supplies nested choices for ValueOrAsk using
// widget.TreeInputOfType and widget.IntConverter. Selected labels are stored as
// an integer slice. This replaces the tree provider and takes precedence over
// flat choices. An empty tree yields an empty slice.
func (p *IntSliceParam) WithTreeOptions(options []widget.TreeOptionOfType[int]) *IntSliceParam {
	return p.WithTreeOptionsFn(func() ([]widget.TreeOptionOfType[int], error) { return options, nil })
}

// WithTreeOptionsFn lazily supplies nested choices for ValueOrAsk. The function
// is called only when the parameter is unset. It replaces any previous tree
// provider; passing nil switches back to flat selection.
func (p *IntSliceParam) WithTreeOptionsFn(fn func() ([]widget.TreeOptionOfType[int], error)) *IntSliceParam {
	p.treeOptionsFn = fn
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
// choose from flat values or tree options. Tree choices are supplied through
// WithTreeOptions or WithTreeOptionsFn; flat choices use WithPossibleValues or
// WithPossibleValuesFn. It fails if no choice source has been set.
//
// After successful selection the chosen slice is stored and the
// parameter is considered set.
func (p *IntSliceParam) ValueOrAsk(ctx context.Context) ([]int, error) {
	if p.IsSet() {
		return p.Value(), nil
	}
	var values []int
	var err error
	if p.treeOptionsFn != nil {
		options, loadErr := p.treeOptionsFn()
		if loadErr != nil {
			return nil, erratt.Errorf("cannot get tree options: %w", loadErr)
		}
		values, err = widget.TreeInputOfType(ctx, p.Description, options, widget.IntConverter)
	} else {
		if len(p.possibleValues) == 0 && p.possibleValuesFn == nil {
			return nil, erratt.New("IntSliceParam ValueOrAsk invoked but possibleValues are not set", "name", p.Name)
		}
		possibleValues := p.possibleValues
		if len(possibleValues) == 0 {
			possibleValues, err = p.possibleValuesFn()
			if err != nil {
				return nil, erratt.Errorf("cannot get possible values: %w", err)
			}
			if len(possibleValues) == 0 {
				return []int{}, nil
			}
		}
		values, err = widget.MultiInputOfType(ctx, p.Description, possibleValues, widget.IntConverter)
	}
	if err != nil {
		return nil, err
	}
	viper.Set(p.Name, values)
	return values, nil
}
