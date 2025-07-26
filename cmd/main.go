package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/vignesh-goutham/AthenaX/cmd/runstrategy"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "athenax",
		Short: "AthenaX - A trading strategy execution engine",
		Long: `AthenaX is a command-line tool for executing trading strategies.
It provides various subcommands to run different trading strategies.`,
	}

	// Add subcommands
	rootCmd.AddCommand(runstrategy.NewRunStrategyCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
