/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"log"

	"github.com/94peter/microservice"
	"github.com/arwoosa/notifaction/router"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the API service based on the configuration",
	Long: `The serve command starts the API service that provides email sending functionality to users.
It initializes the necessary APIs (e.g., notification, health check).
Additionally, it can run a test API for local development to simulate API requests from other microservices.`,
	Run: func(cmd *cobra.Command, args []string) {
		showInfo()
		fmt.Println("serve called", viper.GetString("service"))
		apiServ, err := microservice.NewApiWithViper(microservice.WithAPI(router.GetApis()...))
		if err != nil {
			log.Fatal(err)
			return
		}
		microservice.RunService(apiServ)
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// serveCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// serveCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
