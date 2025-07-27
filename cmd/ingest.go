package cmd

import (
	"fmt"

	"github.com/sandwichlabs/agent-memory-graph/internal/ingest"
	"github.com/spf13/cobra"
)

var ingestCmd = &cobra.Command{
	Use:   "ingest [file path]",
	Short: "Ingest a file into the memory graph",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filePath := args[0]
		err := ingest.IngestFile(filePath)
		if err != nil {
			fmt.Printf("Error ingesting file: %v\n", err)
			return
		}
		fmt.Printf("Ingested file: %s\n", filePath)
	},
}

func init() {
	rootCmd.AddCommand(ingestCmd)
}
