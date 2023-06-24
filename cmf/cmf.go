package cmf

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/yoennisrg/gitflow-cli/template"

	color "github.com/logrusorgru/aurora/v3"
)

const version = "1.0"
const defaultYamlFile = "resources/default.yaml"
const defaultCMFFile = ".gitflow.yaml"

type Repository interface {
	CheckWorkspaceChanges()
	Commit(message string)
	Amend(message string)
	BranchName() string
	NewBranch(message string)
}

type TemplateManager interface {
	Run(template templaterunner.Template, injectedVariables map[string]string) string
}

type FS interface {
	GetFileFromVirtualFS(path string) (string, error)
	GetFileFromFS(path string) (string, error)
	GetCurrentDirectory() (string, error)
	GetCMFile() string
	ParseYaml(template interface{}) error
}

type cmf struct {
	repository      Repository
	templateManager TemplateManager
	fs              FS
}

type CMF interface {
	GetVersion()
	CommitChanges()
	CommitAmend()
	InitializeProject()
	NewBranch()
}

func askForConfirmation(s string) bool {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("%s [y/n]: ", s)

		response, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		response = strings.ToLower(strings.TrimSpace(response))

		if response == "y" || response == "yes" {
			return true
		} else if response == "n" || response == "no" {
			return false
		}
	}
}

func NewCMF(repository Repository, templateManager TemplateManager, fsManager FS) CMF {
	return &cmf{
		repository:      repository,
		templateManager: templateManager,
		fs:              fsManager,
	}
}

func (cmfInstance *cmf) getInnerVariables() map[string]string {
	extra := map[string]string{
		"BRANCH_NAME": cmfInstance.repository.BranchName(),
	}

	return extra
}

// GetVersion return current cmf version
func (cmfInstance *cmf) GetVersion() {
	fmt.Println("Git - Commit Message Formatter v", version)
}

// CommitChanges perform a commit changes over current repository
func (cmfInstance *cmf) CommitChanges() {
	cmfInstance.repository.CheckWorkspaceChanges()

	commitTemplate := templaterunner.CommitTemplate{}
	err := cmfInstance.fs.ParseYaml(&commitTemplate)
	if err != nil {
		fmt.Println("Error parsing YAML:", err)
		return
	}

	message := cmfInstance.templateManager.Run(templaterunner.Template{
		EnvFile:  commitTemplate.EnvFile,
		Env:      commitTemplate.Env,
		Prompt:   commitTemplate.Commit,
		Template: commitTemplate.CommitTemplate,
	}, cmfInstance.getInnerVariables())

	cmfInstance.repository.Commit(message)
}

// CommitAmend perform a commit amend over current repository
func (cmfInstance *cmf) CommitAmend() {
	commitTemplate := templaterunner.CommitTemplate{}

	err := cmfInstance.fs.ParseYaml(&commitTemplate)
	if err != nil {
		fmt.Println("Error parsing YAML:", err)
		return
	}

	message := cmfInstance.templateManager.Run(templaterunner.Template{
		EnvFile:  commitTemplate.EnvFile,
		Env:      commitTemplate.Env,
		Prompt:   commitTemplate.Commit,
		Template: commitTemplate.CommitTemplate,
	}, cmfInstance.getInnerVariables())

	cmfInstance.repository.Amend(message)
}

func (cmfInstance *cmf) NewBranch() {
	branchTemplate := templaterunner.BranchTemplate{}
	err := cmfInstance.fs.ParseYaml(&branchTemplate)
	if err != nil {
		fmt.Println("Error parsing YAML:", err)
		return
	}

	message := cmfInstance.templateManager.Run(templaterunner.Template{
		EnvFile:  branchTemplate.EnvFile,
		Env:      branchTemplate.Env,
		Prompt:   branchTemplate.Branch,
		Template: branchTemplate.BranchTemplate,
	}, cmfInstance.getInnerVariables())

	cmfInstance.repository.NewBranch(message)
}

// InitializeProject initialize current directory with a inner cmf template
func (cmfInstance *cmf) InitializeProject() {
	if askForConfirmation("Create a new .gitflow.yaml file on your working directory. Do you want to continue?") {
		currentDirectory, _ := cmfInstance.fs.GetCurrentDirectory()
		cmfFilePath := currentDirectory + "/" + defaultCMFFile
		cmfFile, _ := cmfInstance.fs.GetFileFromVirtualFS(defaultYamlFile)
		err := ioutil.WriteFile(cmfFilePath, []byte(cmfFile), 0644)
		if err != nil {
			fmt.Println(color.Red("Cannot create .gitflow.yaml file"))
			os.Exit(2)
		}

		fmt.Println(color.Green("Ref: github@yoennisrg/gitflow-cli"))
	}
}
