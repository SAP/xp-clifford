package widget

import (
	"context"
	"strconv"

	"github.com/SAP/xp-clifford/erratt"
	"github.com/charmbracelet/huh"
)

// validatedInput presents an interactive single-line input prompt and returns the entered value as a string.
// It creates a form with a single input field that validates user input according to the provided validator function.
//
// Parameters:
//   - ctx: Context for cancellation or timeout control of the prompt.
//   - title: Text displayed above the input field as a label.
//   - placeholder: Hint text shown inside the input field when it is empty.
//   - validator: Function to validate the input; returns an error if validation fails.
//
// Returns:
//   - The entered string value if successful.
//   - An error if the form was cancelled, timed out, or encountered an issue.
func validatedInput(ctx context.Context, title, placeholder string, validator func(string) error) (string, error) {
	var value string
	echoMode := huh.EchoModeNormal
	input := huh.NewInput().
		Value(&value).
		Title(title).
		Placeholder(placeholder).
		Validate(validator).
		EchoMode(echoMode)
	g := huh.NewGroup(input).WithShowHelp(false)
	f := huh.NewForm(g).WithTheme(huh.ThemeCatppuccin())
	if err := f.RunWithContext(ctx); err != nil {
		return "", err
	}
	return value, nil
}

// intValidator validates that the given string can be parsed as an integer.
//
// Parameters:
//   - s: The string to validate.
//
// Returns:
//   - nil if the string is a valid integer representation.
//   - An annotated error with the invalid input if parsing fails.
func intValidator(s string) error {
	_, err := strconv.Atoi(s)
	if err != nil {
		return erratt.Errorf("invalid input: %w", err).With("input", s)
	}
	return nil
}

// IntInput presents an interactive single-line prompt for entering an integer value.
// The prompt validates user input in real-time and only accepts valid integer numbers.
// The form uses the Catppuccin theme for styling.
//
// Parameters:
//   - ctx: Context for cancellation or timeout control of the prompt. If the context
//     is cancelled or times out, the function returns an error.
//   - title: Text displayed above the input field as a label to describe the expected input.
//   - placeholder: Hint text shown inside the input field when it is empty, providing
//     an example or guidance to the user.
//
// Returns:
//   - The entered integer value if successful.
//   - An error if the form was cancelled, timed out, or if parsing the final value fails.
func IntInput(ctx context.Context, title, placeholder string) (int, error) {
	s, err := validatedInput(ctx, title, placeholder, intValidator)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(s)
}
