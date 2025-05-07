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
var detailTplCmd = &cobra.Command{
	Use:   "detailTpl",
	Short: "Display detailed information about a mail template",
	Long:  `Display detailed information about a mail template, including its title, subject, HTML body, and plain text body.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("listTpl called")
		fmt.Println()
		tplName, err := cmd.Flags().GetString("name")
		errorHandler(err)
		mailTpl, err := factory.NewTemplate()
		errorHandler(err)
		result, err := mailTpl.Detail(tplName)
		errorHandler(err)
		fmt.Println("Template Name: ", result.Title)
		fmt.Println("Template Subject: ", result.Subject)
		fmt.Println("Template Body (HTML): \n", result.Body.Html)
		fmt.Println("Template Body (PLAIN): \n", result.Body.Plaint)

	},
}

func init() {
	mailCmd.AddCommand(detailTplCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// createTplCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// createTplCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	detailTplCmd.Flags().StringP("name", "n", "", "template name")
}
