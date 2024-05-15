/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/mcfe91/go-cli-grpc-chat/client"
	"github.com/spf13/cobra"
)

var remoteServerHost, receivers, name string

// connectCmd represents the connect command
var connectCmd = &cobra.Command{
	Use:   "connect",
	Short: "A brief description of your command",
	Run: func(cmd *cobra.Command, args []string) {
		if name == "" {
			cobra.CheckErr(fmt.Sprintf("please provide your name with `name` argument"))
		}
		client.StartClient(name, receivers, remoteServerHost)
	},
}

func init() {
	rootCmd.AddCommand(connectCmd)

	connectCmd.Flags().StringVarP(&remoteServerHost, "remove-server-host", "s", "localhost:3000", "Remote server host where you want to join chat e.g 10.11.12.13:8080, default is localhost:3000")
	connectCmd.Flags().StringVarP(&receivers, "chatting-with", "c", "all", "comma separated list of users names on the remote host you want to chat with e.g. A,B,C, default is you can chat with all")
	connectCmd.Flags().StringVarP(&name, "name", "n", "", "your display name you want users to see e.g. Piyush")
}
