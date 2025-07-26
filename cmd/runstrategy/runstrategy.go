package runstrategy

import (
	"context"
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/vignesh-goutham/AthenaX/pkg/alpaca"
	"github.com/vignesh-goutham/AthenaX/pkg/engine"
	"github.com/vignesh-goutham/AthenaX/pkg/notification"
	"github.com/vignesh-goutham/AthenaX/pkg/strategies"
)

var (
	strategyName string
)

// NewRunStrategyCmd creates the run-strategy command
func NewRunStrategyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run-strategy",
		Short: "Run a specific trading strategy",
		Long: `Run a specific trading strategy by name.
Available strategies:
- two-percent-down: Executes the 2% gap down strategy`,
		RunE: runStrategy,
	}

	// Add flags
	cmd.Flags().StringVarP(&strategyName, "name", "n", "", "Name of the strategy to run (required)")
	cmd.MarkFlagRequired("name")

	return cmd
}

func runStrategy(cmd *cobra.Command, args []string) error {
	// Create broker client
	broker, err := alpaca.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create broker client: %w", err)
	}

	// Create notification client
	notifier, err := notification.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create notification client: %w", err)
	}

	// Create strategy based on name
	var strategy strategies.Strategy
	switch strategyName {
	case "two-percent-down":
		strategy = strategies.NewTwoPercentDown(broker, notifier)
	default:
		return fmt.Errorf("unknown strategy: %s", strategyName)
	}

	// Create engine with the strategy
	eng := engine.NewEngine([]strategies.Strategy{strategy}, broker, notifier)

	// Create context
	ctx := context.Background()

	log.Printf("Running strategy: %s", strategyName)

	// Run the engine
	if err := eng.Run(ctx); err != nil {
		return fmt.Errorf("failed to run strategy: %w", err)
	}

	log.Printf("Strategy %s completed successfully", strategyName)
	return nil
}
