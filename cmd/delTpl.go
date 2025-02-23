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
var delTplCmd = &cobra.Command{
	Use:   "delTpl",
	Short: "Delete an email template from AWS SES",
	Long: `Deletes a specified email template from AWS SES.
This command checks if the given template exists; if so, it removes it.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("delTpl called")
		name, err := cmd.Flags().GetString("name")
		errorHandler(err)
		if name == "" {
			fmt.Println("name is required")
			return
		}
		mailTpl, err := factory.NewTemplate()
		errorHandler(err)
		err = mailTpl.Delete(name)
		errorHandler(err)
		fmt.Println("success")
	},
}

func init() {
	mailCmd.AddCommand(delTplCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// createTplCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// createTplCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	delTplCmd.Flags().StringP("name", "n", "", "name")
}
