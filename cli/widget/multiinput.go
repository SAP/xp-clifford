package widget

import (
	"context"
	"fmt"
	"strconv"

	"github.com/SAP/xp-clifford/erratt"
	"github.com/charmbracelet/huh"
)

func makeSelectOption(options []string) []huh.Option[string] {
	selects := make([]huh.Option[string], len(options))
	for i := range options {
		selects[i] = huh.NewOption(options[i], options[i])
	}
	return selects
}

func multiInputdescFuncFn(items int, selected *[]string) func() string {
	return func() string {
		return fmt.Sprintf("%d/%d selected", len(*selected), items)
	}
}

// MultiInput presents an interactive multi-select widget and returns
// the chosen options.
//
// ctx is used to cancel or time-out the prompt.
// title is displayed above the list.
// options supplies the selectable items.
func MultiInput(ctx context.Context, title string, options []string) ([]string, error) {
	selected := []string{}
	multiSelect := huh.NewMultiSelect[string]().
		Options(
			makeSelectOption(options)...,
		).
		Description(multiInputdescFuncFn(len(options), &selected)()).
		DescriptionFunc(multiInputdescFuncFn(len(options), &selected), &selected).
		Title(title).
		Value(&selected)
	g := huh.NewGroup(multiSelect).WithShowHelp(true)
	f := huh.NewForm(g).WithTheme(huh.ThemeCatppuccin()).WithHeight(10)
	if err := f.RunWithContext(ctx); err != nil {
		return nil, err
	}
	return selected, nil
}

type TypeConverter[T any] interface {
	FromType(T) (string, error)
	ToType(string) (T, error)
}

type intConverterType struct{}

func (c intConverterType) FromType(i int) (string, error) {
	return strconv.Itoa(i), nil
}

func (c intConverterType) ToType(s string) (int, error) {
	i, err := strconv.ParseInt(s, 10, strconv.IntSize)
	if err != nil {
		return 0, err
	}
	return int(i), nil
}

var IntConverter TypeConverter[int] = intConverterType{}

func MultiInputOfType[T any](ctx context.Context, title string, options []T, converter TypeConverter[T]) ([]T, error) {
	in := make([]string, len(options))
	for i := range in {
		var err error
		in[i], err = converter.FromType(options[i])
		if err != nil {
			return nil, erratt.Errorf("error converting to string: %w", err).With("value", in[i])
		}
	}
	outStrings, err := MultiInput(ctx, title, in)
	if err != nil {
		return nil, err
	}
	out := make([]T, len(outStrings))
	for i := range outStrings {
		out[i], err = converter.ToType(outStrings[i])
		if err != nil {
			return nil, erratt.Errorf("error converting from string: %w", err).With("value", out[i])
		}
	}
	return out, nil
}
