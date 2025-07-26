package main

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/vignesh-goutham/AthenaX/pkg/alpaca"
	"github.com/vignesh-goutham/AthenaX/pkg/engine"
	"github.com/vignesh-goutham/AthenaX/pkg/notification"
	"github.com/vignesh-goutham/AthenaX/pkg/strategies"
)

// LambdaEvent represents the input event for the Lambda function
type LambdaEvent struct {
	StrategyName string `json:"strategy_name"`
	// Add other fields as needed for your use case
}

// LambdaResponse represents the response from the Lambda function
type LambdaResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// Handler is the main Lambda function handler
func Handler(ctx context.Context, event LambdaEvent) (LambdaResponse, error) {
	log.Printf("Received event: %+v", event)

	// Validate strategy name
	if event.StrategyName == "" {
		return LambdaResponse{
			Status:  "error",
			Message: "Strategy name is required",
			Error:   "strategy_name is empty",
		}, nil
	}

	// Create broker client
	broker, err := alpaca.NewClient()
	if err != nil {
		log.Printf("Failed to create broker client: %v", err)
		return LambdaResponse{
			Status:  "error",
			Message: "Failed to create broker client",
			Error:   err.Error(),
		}, nil
	}

	// Create notification client
	notifier, err := notification.NewClient()
	if err != nil {
		log.Printf("Failed to create notification client: %v", err)
		return LambdaResponse{
			Status:  "error",
			Message: "Failed to create notification client",
			Error:   err.Error(),
		}, nil
	}

	// Create strategy based on name
	var strategy strategies.Strategy
	switch event.StrategyName {
	case "two-percent-down":
		strategy = strategies.NewTwoPercentDown(broker, notifier)
	default:
		return LambdaResponse{
			Status:  "error",
			Message: fmt.Sprintf("Unknown strategy: %s", event.StrategyName),
			Error:   "unknown strategy",
		}, nil
	}

	// Create engine with the strategy
	eng := engine.NewEngine([]strategies.Strategy{strategy}, broker, notifier)

	log.Printf("Running strategy: %s", event.StrategyName)

	// Run the engine
	if err := eng.Run(ctx); err != nil {
		log.Printf("Failed to run strategy: %v", err)
		return LambdaResponse{
			Status:  "error",
			Message: "Failed to run strategy",
			Error:   err.Error(),
		}, nil
	}

	log.Printf("Strategy %s completed successfully", event.StrategyName)
	return LambdaResponse{
		Status:  "success",
		Message: fmt.Sprintf("Strategy %s completed successfully", event.StrategyName),
	}, nil
}

func main() {
	lambda.Start(Handler)
}
