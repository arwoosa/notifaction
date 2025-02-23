/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"time"

	"github.com/arwoosa/notifaction/service/mail/factory"
	"github.com/spf13/cobra"
)

// createTplCmd represents the createTpl command
var listTplCmd = &cobra.Command{
	Use:   "listTpl",
	Short: "List email templates from AWS SES",
	Long: `Lists all email templates stored in AWS SES with their creation times.
Supports pagination via the --next-token flag, which allows retrieval of subsequent results
if the number of templates exceeds a single page.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("listTpl called")
		fmt.Println()
		nextToken, err := cmd.Flags().GetString("next-token")
		errorHandler(err)
		mailTpl, err := factory.NewTemplate()
		errorHandler(err)
		result, err := mailTpl.List(nextToken)
		errorHandler(err)
		fmt.Println("Template Name                     Created Time")
		fmt.Println("========================================================")
		for _, t := range result.Templates {
			fmt.Printf("%-30s    %s\n", t.Name, t.CreateTime.Format(time.RFC3339))
		}
		if result.NextToken != nil {
			fmt.Printf("next token: %s\n", *result.NextToken)
		}
	},
}

func init() {
	mailCmd.AddCommand(listTplCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// createTplCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// createTplCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	listTplCmd.Flags().StringP("next-token", "t", "", "next token")
}
