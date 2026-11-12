package configparam

import (
	"context"

	"github.com/SAP/xp-clifford/cli/widget"
	"github.com/SAP/xp-clifford/erratt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type stringSlicePossibleValuesFnType func() ([]string, error)
type stringSliceTreeOptionsFnType func() ([]widget.TreeOption, error)

// StringSliceParam type represents a configuration parameterkxo that
// holds a []string value.
type StringSliceParam struct {
	*configWithDefaultValue[StringSliceParam, []string]
	sensitive        bool
	possibleValues   []string
	possibleValuesFn stringSlicePossibleValuesFnType
	treeOptionsFn    stringSliceTreeOptionsFnType
}

var _ ConfigParam = &StringSliceParam{}

func stringSliceGenerator(name, description string, sensitive bool) *StringSliceParam {
	p := &StringSliceParam{
		sensitive:      sensitive,
		possibleValues: []string{},
	}
	p.configWithDefaultValue = newConfigWithDefaultValue(p, name, description, []string{})
	return p
}

// StringSlice creates a [StringSliceParam] value. The mandatory name
// and description parameters must be set.
func StringSlice(name, description string) *StringSliceParam {
	return stringSliceGenerator(name, description, false)
}

// SensitiveStringSlice returns a new [StringSliceParam] whose
// contents are masked when the parameter is printed to the
// console. Both name and description are required.
func SensitiveStringSlice(name, description string) *StringSliceParam {
	return stringSliceGenerator(name, description, true)
}

// WithPossibleValues restricts the interactive selection to the
// supplied slice. When the CLI prompts the user, only these strings
// are offered as choices. This switches to flat selection.
func (p *StringSliceParam) WithPossibleValues(values []string) *StringSliceParam {
	p.possibleValues = values
	p.treeOptionsFn = nil
	return p
}

// WithPossibleValuesFn lazily supplies the valid choices for
// interactive selection. The given function is called when the CLI
// prompts the user. The returned strings are presented as options.
// This switches to flat selection.
func (p *StringSliceParam) WithPossibleValuesFn(fn func() ([]string, error)) *StringSliceParam {
	p.possibleValuesFn = fn
	p.treeOptionsFn = nil
	return p
}

// WithTreeOptions supplies nested choices for ValueOrAsk using widget.TreeInput.
// Selected labels are stored as a string slice. This replaces the tree provider
// and takes precedence over flat choices. An empty tree yields an empty slice.
func (p *StringSliceParam) WithTreeOptions(options []widget.TreeOption) *StringSliceParam {
	return p.WithTreeOptionsFn(func() ([]widget.TreeOption, error) { return options, nil })
}

// WithTreeOptionsFn lazily supplies nested choices for ValueOrAsk. The function
// is called only when the parameter is unset. It replaces any previous tree
// provider; passing nil switches back to flat selection.
func (p *StringSliceParam) WithTreeOptionsFn(fn func() ([]widget.TreeOption, error)) *StringSliceParam {
	p.treeOptionsFn = fn
	return p
}

// AttachToCommand registers the persistent string-slice flag (long form and
// optional short form) with the supplied [cobra.Command].
func (p *StringSliceParam) AttachToCommand(command *cobra.Command) {
	p.attachToCommand(command.PersistentFlags().StringSlice, command.PersistentFlags().StringSliceP)
}

// Value returns the user configured string slice. If the user has not
// configured any values, the default value is returned.
func (p *StringSliceParam) Value() []string {
	return p.value(viper.GetStringSlice)
}

// ValueOrAsk returns the configured slice or prompts the user to
// choose from flat values or tree options. Tree choices are supplied through
// WithTreeOptions or WithTreeOptionsFn; flat choices use WithPossibleValues or
// WithPossibleValuesFn. It fails if no choice source has been set.
//
// After successful selection the chosen slice is stored and the
// parameter is considered set.
func (p *StringSliceParam) ValueOrAsk(ctx context.Context) ([]string, error) {
	if p.IsSet() {
		return p.Value(), nil
	}
	var values []string
	var err error
	if p.treeOptionsFn != nil {
		options, loadErr := p.treeOptionsFn()
		if loadErr != nil {
			return nil, erratt.Errorf("cannot get tree options: %w", loadErr)
		}
		values, err = widget.TreeInput(ctx, p.Description, options)
	} else {
		if len(p.possibleValues) == 0 && p.possibleValuesFn == nil {
			return nil, erratt.New("StringSliceParam ValueOrAsk invoked but possibleValues are not set", "name", p.Name)
		}
		possibleValues := p.possibleValues
		if len(possibleValues) == 0 {
			possibleValues, err = p.possibleValuesFn()
			if err != nil {
				return nil, erratt.Errorf("cannot get possible values: %w", err)
			}
			if len(possibleValues) == 0 {
				return []string{}, nil
			}
		}
		values, err = widget.MultiInput(ctx, p.Description, possibleValues)
	}
	if err != nil {
		return nil, err
	}
	viper.Set(p.Name, values)
	return values, nil
}
