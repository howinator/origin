package main

import (
	"os"

	"github.com/howinator/house/cmd/adblock"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "house",
		Short: "CLI for managing home infrastructure",
	}
	rootCmd.AddCommand(adblock.Cmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
