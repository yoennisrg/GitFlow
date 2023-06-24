package cmd

import (
	"embed"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/yoennisrg/gitflow-cli/cmf"
	"github.com/yoennisrg/gitflow-cli/fs"
	git "github.com/yoennisrg/gitflow-cli/git"
	promptxwrapper "github.com/yoennisrg/gitflow-cli/promptxWrapper"
	"github.com/yoennisrg/gitflow-cli/template"
)

var cmfInstance cmf.CMF

// Root Root cli command
var root = &cobra.Command{
	Use:   "gitflow",
	Short: "GitFlow CLI",
	Long:  "Utility for standardizing project confirmation messages and temporary branches.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Tool under development please read the readme")
	},
}

var config = &cobra.Command{
	Use:   "init",
	Short: "Create configuration file",
	Long:  "Create .gitflow.yaml configuration file",
	Run: func(cmd *cobra.Command, args []string) {
		cmfInstance.InitializeProject()
	},
}

// var amend = &cobra.Command{
// 	Use:   "amend",
// 	Short: "Amend commit message",
// 	Long:  "Amend last commit message",
// 	Run: func(cmd *cobra.Command, args []string) {
// 		cmfInstance.CommitAmend()
// 	},
// }

var add = &cobra.Command{
	Use:   "add",
	Short: "Commit message",
	Long:  "Customized commit messages by steps",
	Run: func(cmd *cobra.Command, args []string) {
		cmfInstance.CommitChanges()
	},
}

var check = &cobra.Command{
	Use:   "check",
	Short: "Checkout branch",
	Long:  "Customized branch name temportal by steps",
	Run: func(cmd *cobra.Command, args []string) {
		cmfInstance.NewBranch()
	},
}

func Execute() {
	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func Build(vfs embed.FS) {
	fsManager := fs.NewFs(vfs)
	promptManager := promptxwrapper.NewPromptxWrapper()
	templateManager := templaterunner.NewTemplateRunner(promptManager)
	cmfInstance = cmf.NewCMF(git.NewGitWrapper(), templateManager, fsManager)
	// root.AddCommand(version)
	root.AddCommand(config)
	// root.AddCommand(amend)
	root.AddCommand(add)
	root.AddCommand(check)
}
