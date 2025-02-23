/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// mailCmd represents the mail command
var mailCmd = &cobra.Command{
	Use:   "mail",
	Short: "Manage AWS SES email templates",
	Long: `The mail command provides a set of subcommands to manage email templates in AWS SES.
It supports applying templates from a YAML file (applyTpl), deleting existing templates (delTpl),
and listing all stored templates with pagination support (listTpl). Use these subcommands to seamlessly
create, update, remove, or query email templates in your AWS SES environment.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("mail called")
	},
}

func init() {
	rootCmd.AddCommand(mailCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// mailCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// mailCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
