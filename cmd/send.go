package cmd

import (
	"fmt"
	"strings"

	"github.com/arwoosa/notifaction/service"
	"github.com/arwoosa/notifaction/service/mail/factory"
	"github.com/spf13/cobra"
)

var sendCmd = &cobra.Command{
	Use:   "send",
	Short: "Send an email notification",
	Long: `Send an email notification using the configured mail provider.

Example:
  # Send an email with template "welcome_email" in English
  notifaction mail send --event welcome_email --lang en --to user@example.com --data "name=John"

  # Send an email with multiple recipients and data
  notifaction mail send --event welcome_email --lang en --to user1@example.com --to user2@example.com --data "name=John&age=30"
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		tplname, _ := cmd.Flags().GetString("n")
		to, _ := cmd.Flags().GetStringSlice("to")
		data, _ := cmd.Flags().GetString("data")

		if tplname == "" {
			return fmt.Errorf("-n is required")
		}

		splitIdx := strings.LastIndex(tplname, "_")
		if splitIdx == -1 {
			return fmt.Errorf("invalid template name: %s", tplname)
		}

		event := tplname[:splitIdx]
		lang := tplname[splitIdx+1:]

		notification := &service.Notification{
			Event:  event,
			Lang:   lang,
			SendTo: make([]*service.Info, len(to)),
			Data:   make(map[string]string),
		}

		if len(to) == 0 {
			return fmt.Errorf("--to is required")
		}

		// Parse data string into map
		if data != "" {
			pairs := strings.Split(data, "&")
			for _, pair := range pairs {
				keyValue := strings.SplitN(pair, "=", 2)
				if len(keyValue) == 2 {
					notification.Data[keyValue[0]] = keyValue[1]
				}
			}
		}

		// Create Info objects for recipients
		for i, email := range to {
			notification.SendTo[i] = &service.Info{
				Email: email,
			}
		}

		sender, err := factory.NewApiSender()
		if err != nil {
			return fmt.Errorf("failed to create mail sender: %w", err)
		}

		messageId, err := sender.Send(notification)
		if err != nil {
			return fmt.Errorf("failed to send email: %w", err)
		}

		fmt.Printf("Email sent successfully. Message ID: %s\n", messageId)
		return nil
	},
}

func init() {
	mailCmd.AddCommand(sendCmd)

	sendCmd.Flags().String("n", "", "Template name")
	sendCmd.Flags().StringSlice("to", []string{}, "Recipient email address (can be specified multiple times)")
	sendCmd.Flags().String("data", "", "Template data in key=value format (multiple values can be separated by &)\nExample: name=John&age=30")
	sendCmd.MarkFlagRequired("n")
	sendCmd.MarkFlagRequired("to")
}
