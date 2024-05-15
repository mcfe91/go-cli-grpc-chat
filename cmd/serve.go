/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/mcfe91/go-cli-grpc-chat/server"
	"github.com/spf13/cobra"
)

var port string

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "serve starts the chat server",
	Run: func(cmd *cobra.Command, args []string) {
		server.StartServer(port)
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)

	serveCmd.Flags().StringVarP(&port, "port", "p", "3000", "port on which you want to start the server, default :3000")
}
