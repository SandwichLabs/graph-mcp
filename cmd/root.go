package cmd

import (
	"fmt"
	"os"

	"github.com/sandwichlabs/agent-memory-graph/internal/server"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "amg [Path to Memory Graph Directory]",
	Short: "A CLI to extend MCP with graph data.",
	Long:  `amg is a command-line tool that exposes memory management and knowledge retrieval functions for MCP.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.Help()
			return
		}
		servername, _ := cmd.Flags().GetString("name")
		if servername == "" {
			servername = "knowledge"
		}

		server.Run(args[0], servername)
	},
}

func init() {
	rootCmd.Flags().String("name", "", "Name of the MCP server (default: 'tasks')")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
