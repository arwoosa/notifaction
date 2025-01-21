/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/arwoosa/notifaction/service/mail"
	"github.com/spf13/cobra"
)

// createTplCmd represents the createTpl command
var applyTplCmd = &cobra.Command{
	Use:   "applyTpl",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("applyTpl called")
		file, err := cmd.Flags().GetString("file")
		errorHandler(err)
		mailTpl, err := mail.NewTemplateWithAWS()
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
