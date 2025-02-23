/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/arwoosa/notifaction/service/mail/factory"
	"github.com/spf13/cobra"
)

// createTplCmd represents the createTpl command
var applyTplCmd = &cobra.Command{
	Use:   "applyTpl",
	Short: "Apply an email template from a YAML file to AWS SES",
	Long: `Reads a specified YAML file, validates the email template, 
and applies it to AWS SES. If the template already exists, it will be updated; 
otherwise, a new template will be created.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("applyTpl called")
		file, err := cmd.Flags().GetString("file")
		errorHandler(err)
		mailTpl, err := factory.NewTemplate()
		errorHandler(err)
		err = mailTpl.Apply(file)
		errorHandler(err)
		fmt.Println("success")
	},
}

func init() {
	mailCmd.AddCommand(applyTplCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// createTplCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// createTplCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	applyTplCmd.Flags().StringP("file", "f", "", "template file (YAML)")
}
