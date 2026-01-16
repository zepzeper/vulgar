package cli

import (
	"fmt"

	"github.com/charmbracelet/huh"
)

// Select prompts the user to select from a list of options
// Returns the selected value and any error
func Select[T comparable](title string, options []T, display func(T) string) (T, error) {
	var selected T

	if len(options) == 0 {
		return selected, fmt.Errorf("no options provided")
	}

	if len(options) == 1 {
		return options[0], nil
	}

	huhOptions := make([]huh.Option[T], len(options))
	for i, opt := range options {
		label := display(opt)
		huhOptions[i] = huh.NewOption(label, opt)
	}

	err := huh.NewSelect[T]().
		Title(title).
		Options(huhOptions...).
		Value(&selected).
		WithTheme(CustomHuhTheme()).
		Run()

	if err != nil {
		return selected, fmt.Errorf("selection cancelled")
	}

	return selected, nil
}

func SelectString(title string, options []string) (string, error) {
	return Select(title, options, func(s string) string { return s })
}

// SelectWithDefault prompts the user to select with a default value
func SelectWithDefault[T comparable](title string, options []T, defaultVal T, display func(T) string) (T, error) {
	var selected T = defaultVal

	if len(options) == 0 {
		return selected, fmt.Errorf("no options provided")
	}

	if len(options) == 1 {
		return options[0], nil
	}

	huhOptions := make([]huh.Option[T], len(options))
	for i, opt := range options {
		label := display(opt)
		huhOptions[i] = huh.NewOption(label, opt)
	}

	err := huh.NewSelect[T]().
		Title(title).
		Options(huhOptions...).
		Value(&selected).
		WithTheme(CustomHuhTheme()).
		Run()

	if err != nil {
		return defaultVal, fmt.Errorf("selection cancelled")
	}

	return selected, nil
}

func Confirm(title string, defaultVal bool) (bool, error) {
	var confirmed bool = defaultVal

	err := huh.NewConfirm().
		Title(title).
		Affirmative("Yes").
		Negative("No").
		Value(&confirmed).
		WithTheme(CustomHuhTheme()).
		Run()

	if err != nil {
		return defaultVal, fmt.Errorf("confirmation cancelled")
	}

	return confirmed, nil
}

func Input(title string, placeholder string) (string, error) {
	var value string

	err := huh.NewInput().
		Title(title).
		Placeholder(placeholder).
		Value(&value).
		WithTheme(CustomHuhTheme()).
		Run()

	if err != nil {
		return "", fmt.Errorf("input cancelled")
	}

	return value, nil
}

func InputWithDefault(title string, defaultVal string) (string, error) {
	value := defaultVal

	err := huh.NewInput().
		Title(title).
		Value(&value).
		Run()

	if err != nil {
		return defaultVal, fmt.Errorf("input cancelled")
	}

	return value, nil
}

func InputPassword(title string) (string, error) {
	var value string

	err := huh.NewInput().
		Title(title).
		EchoMode(huh.EchoModePassword).
		Value(&value).
		Run()

	if err != nil {
		return "", fmt.Errorf("input cancelled")
	}

	return value, nil
}

func MultiSelect[T comparable](title string, options []T, display func(T) string) ([]T, error) {
	var selected []T

	if len(options) == 0 {
		return selected, fmt.Errorf("no options provided")
	}

	huhOptions := make([]huh.Option[T], len(options))
	for i, opt := range options {
		label := display(opt)
		huhOptions[i] = huh.NewOption(label, opt)
	}

	err := huh.NewMultiSelect[T]().
		Title(title).
		Options(huhOptions...).
		Value(&selected).
		WithTheme(CustomHuhTheme()).
		Run()

	if err != nil {
		return nil, fmt.Errorf("selection cancelled")
	}

	return selected, nil
}

func MultiSelectString(title string, options []string) ([]string, error) {
	return MultiSelect(title, options, func(s string) string { return s })
}

func TextArea(title string, placeholder string) (string, error) {
	var value string

	err := huh.NewText().
		Title(title).
		Placeholder(placeholder).
		Value(&value).
		WithTheme(CustomHuhTheme()).
		Run()

	if err != nil {
		return "", fmt.Errorf("input cancelled")
	}

	return value, nil
}
