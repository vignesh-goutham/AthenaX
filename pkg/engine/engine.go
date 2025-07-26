package engine

import (
	"context"

	"github.com/vignesh-goutham/AthenaX/pkg/alpaca"
	"github.com/vignesh-goutham/AthenaX/pkg/strategies"
)

type Engine struct {
	strategies []strategies.Strategy
	broker     *alpaca.Client
}

func NewEngine(strategies []strategies.Strategy, broker *alpaca.Client) *Engine {
	return &Engine{
		strategies: strategies,
		broker:     broker,
	}
}

func (e *Engine) Run(ctx context.Context) error {
	// Check if market is open first
	// isOpen, err := e.broker.IsMarketOpen(ctx)
	// if err != nil {
	// 	return err
	// }
	// if !isOpen {
	// 	log.Println("Market is closed, exiting...")
	// 	return nil
	// }

	// Run strategies only if market is open
	for _, strategy := range e.strategies {
		if err := strategy.Run(ctx); err != nil {
			return err
		}
	}
	return nil
}
