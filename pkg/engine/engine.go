package engine

import (
	"context"
	"fmt"
	"log"

	"github.com/vignesh-goutham/AthenaX/pkg/alpaca"
	"github.com/vignesh-goutham/AthenaX/pkg/notification"
	"github.com/vignesh-goutham/AthenaX/pkg/strategies"
)

type Engine struct {
	strategies []strategies.Strategy
	broker     *alpaca.Client
	notifier   *notification.Client
}

func NewEngine(strategies []strategies.Strategy, broker *alpaca.Client, notifier *notification.Client) *Engine {
	return &Engine{
		strategies: strategies,
		broker:     broker,
		notifier:   notifier,
	}
}

func (e *Engine) Run(ctx context.Context) error {
	// Check if market is open first
	isOpen, err := e.broker.IsMarketOpen(ctx)
	if err != nil {
		return e.notifier.Failure(fmt.Sprintf("failed to check if market is open: %w", err))
	}
	if !isOpen {
		log.Println("Market is closed, exiting...")
		return e.notifier.MarketClosed()
	}

	// Run strategies only if market is open
	for _, strategy := range e.strategies {
		if err := strategy.Run(ctx); err != nil {
			return err
		}
	}
	return nil
}
