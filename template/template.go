package templaterunner

import (
	"errors"
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"os"
	"regexp"
	"strings"

	"gopkg.in/yaml.v2"
)

type templateRunner struct {
	promptManager PromptManager
}

// PromptManager ...
type PromptManager interface {
	ReadValue(title string, errorMessage string, defaultValue string) string
	ReadValueFromList(title string, options []Options) (string, int)
}

// TemplateRunner main template interface
type TemplateRunner interface {
	Run(template Template, injectedVariables map[string]string) string
}

type CommitTemplate struct {
	EnvFile        []string          `yaml:"ENV_FILE"`
	Env            map[string]string `yaml:"ENV"`
	Commit         []PromptItem      `yaml:"COMMIT"`
	CommitTemplate string            `yaml:"COMMIT_TEMPLATE"`
}

type BranchTemplate struct {
	EnvFile        []string          `yaml:"ENV_FILE"`
	Env            map[string]string `yaml:"ENV"`
	Branch         []PromptItem      `yaml:"BRANCH"`
	BranchTemplate string            `yaml:"BRANCH_TEMPLATE"`
}

// Template main template struct
type Template struct {
	EnvFile  []string          `yaml:"ENV_FILE"`
	Env      map[string]string `yaml:"ENV"`
	Prompt   []PromptItem      `yaml:"COMMIT"`
	Template string            `yaml:"TEMPLATE"`
}

//PromptItem ...
type PromptItem struct {
	Key          string    `yaml:"KEY"`
	Label        string    `yaml:"LABEL"`
	Separator    string    `yaml:"SEPARATOR"`
	ErrorLabel   string    `yaml:"ERROR_LABEL"`
	DefaultValue string    `yaml:"DEFAULT_VALUE"`
	Regex        string    `yaml:"REGEX"`
	Options      []Options `yaml:"OPTIONS"`
	Inputs       []Options `yaml:"INPUTS"`
	ProjectName  string    `yaml:"PROJECT_NAME"`
}

// Options multiselect option struct

type Options struct {
	Value       string `yaml:"VALUE"`
	Description string `yaml:"DESC"`
}

type keyValue struct {
	Key   string
	Value string
}

// NewTemplateRunner return a  bluetnew instance of template
func NewTemplateRunner(promptManager PromptManager) TemplateRunner {
	return &templateRunner{
		promptManager: promptManager,
	}
}

func (tr *templateRunner) parseYaml(yamlData string) (Template, error) {
	template := Template{}
	err := yaml.Unmarshal([]byte(yamlData), &template)

	if err != nil {
		return Template{}, errors.New("parsing yaml error")
	}
	return template, nil
}

// Run return the result of run the template
func (tr *templateRunner) Run(template Template, defaultVariables map[string]string) string {
	variables := []keyValue{}
	for k, v := range defaultVariables {
		variables = append(variables, keyValue{Key: k, Value: v})
	}

	if len(template.EnvFile) > 0 {
		err := godotenv.Load(".local.env")
		if err != nil {
			log.Fatalf("Error loading .local.env file")
		}

		for _, environmentVariable := range template.EnvFile {
			variables = append(variables, keyValue{Key: environmentVariable, Value: os.Getenv(environmentVariable)})
		}
	}

	for envKey, envVal := range template.Env {
		variables = append(variables, keyValue{Key: envKey, Value: envVal})
	}

	promptVariables := tr.prompt(template, variables)
	variables = append(variables, promptVariables...)

	message := tr.parseTemplate(template.Template, variables)

	return message
}

func (tr *templateRunner) parseTemplate(template string, variables []keyValue) string {
	for _, v := range variables {
		template = strings.Replace(template, "{{"+v.Key+"}}", v.Value, -1)
	}

	return template
}

func (tr *templateRunner) prompt(template Template, defaultVariables []keyValue) []keyValue {
	variables := []keyValue{}
	for _, step := range template.Prompt {
		result := ""
		defaultValue := ""
		var errorMessage = "empty value"

		if step.ErrorLabel != "" {
			errorMessage = step.ErrorLabel
		}

		//fmt.Println("stepLabel:", step.Label)
		//fmt.Println("stepKey:", step.Key)
		//fmt.Println("stepRegex:", step.Regex)
		//fmt.Println("stepDefaultValue:", step.DefaultValue)
		//fmt.Println("stepErrorLabel:", step.ErrorLabel)
		//fmt.Println("stepOptionsNil:", step.Options == nil)
		//fmt.Println("stepInputsNil:", step.Inputs == nil)

		if step.Inputs != nil {
			t, position := tr.promptManager.ReadValueFromList(step.Label, step.Inputs)
			currentStep := step.Inputs[position]

			if t == "diff" {

			} else if t == "comment" {
				newStep := PromptItem{
					Key:   currentStep.Value,
					Label: currentStep.Description,
				}
				text := ReadInput(newStep, defaultValue, tr, defaultVariables, result, errorMessage)
				fmt.Println(text)

			} else {
				newStep := PromptItem{
					Key:   currentStep.Value,
					Label: currentStep.Description,
				}
				result = ReadInput(newStep, defaultValue, tr, defaultVariables, result, errorMessage)
			}

		} else if step.Options == nil {
			result = ReadInput(step, defaultValue, tr, defaultVariables, result, errorMessage)
		} else {
			result, _ = tr.promptManager.ReadValueFromList(step.Label, step.Options)
		}

		if step.ProjectName != "" {
			variables = append(variables, keyValue{
				Key:   "PROJECT_NAME",
				Value: tr.parseTemplate(step.ProjectName, defaultVariables),
			})
		}

		if step.Separator != "" {
			result = strings.Replace(result, " ", step.Separator, -1)
		}

		variables = append(variables, keyValue{
			Key:   step.Key,
			Value: result,
		})

		//fmt.Println("variables", variables)
	}

	return variables
}

func ReadInput(step PromptItem, defaultValue string, tr *templateRunner, defaultVariables []keyValue, result string, errorMessage string) string {
	var labelMessage = step.Label
	var projectName = ""

	if step.ProjectName != "" {
		projectName = tr.parseTemplate(step.ProjectName, defaultVariables)
		labelMessage += " (" + projectName + ")"
	}

	if step.DefaultValue != "" {
		defaultValue = tr.parseTemplate(step.DefaultValue, defaultVariables)
		if step.Regex != "" {
			r, _ := regexp.Compile(step.Regex)
			defaultValue = r.FindStringSubmatch(defaultValue)[0]
		}

		labelMessage += " (" + defaultValue + ")"
	}

	labelMessage += ":"

	result = tr.promptManager.ReadValue(labelMessage, errorMessage, defaultValue)

	//if projectName != "" && result != projectName {
	//	return fmt.Sprintf("#%s-%s", projectName, result)
	//}
	return result
}
