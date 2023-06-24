package promptxwrapper

import (
	"errors"
	"strings"

	"github.com/mritd/promptx"
	"github.com/yoennisrg/gitflow-cli/prompt"
	"github.com/yoennisrg/gitflow-cli/template"
)

// PromptxWrapper promptx wrapper object
type PromptxWrapper struct{}

// NewPromptxWrapper return a new instance of promptxWrapper
func NewPromptxWrapper() *PromptxWrapper {
	return &PromptxWrapper{}
}

// ReadValue return a value from user single input
func (pw *PromptxWrapper) ReadValue(title string, errorMessage string, defaultValue string) string {
	input := promptx.NewDefaultPrompt(func(line []rune) error {
		if strings.TrimSpace(string(line)) == "" && defaultValue == "" {
			return errors.New(errorMessage)
		}

		return nil
	}, title)

	value := input.Run()

	if value == "" && defaultValue != "" {
		return defaultValue
	}

	return value
}

func (pw *PromptxWrapper) ReadValueAskGPT(title string) string {

	return ""
}

// ReadValueFromList return a value from user multi select input
func (pw *PromptxWrapper) ReadValueFromList(title string, options []templaterunner.Options) (string, int) {
	configuration := &promptx.SelectConfig{
		ActiveTpl:    "\U0001F449  {{ .Title | cyan | bold }}",
		InactiveTpl:  "    {{ .Title | white }}",
		SelectPrompt: title,
		SelectedTpl:  "\U0001F44D {{ \"" + title + "\" | cyan }} {{ .Title | cyan | bold }}",
		DetailsTpl: `----------------------
    {{ .Title | white | bold }} {{ .Description | white | bold }}`,
	}

	var items []prompt.SelectItem
	for _, option := range options {
		items = append(items, prompt.SelectItem{Title: option.Value, Value: option.Value, Description: option.Description})
	}

	selector := &promptx.Select{
		Items:  items,
		Config: configuration,
	}

	position := selector.Run()
	return items[position].Value, position
}
