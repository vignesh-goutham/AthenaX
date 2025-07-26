package alpaca

import (
	"context"
	"fmt"
	"log"

	"github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
)

// GetAllPositions retrieves all positions in the account
func (c *Client) GetAllPositions(ctx context.Context) ([]alpaca.Position, error) {
	positions, err := c.tradingClient.GetPositions()
	if err != nil {
		return nil, fmt.Errorf("failed to get positions: %w", err)
	}

	return positions, nil
}

// GetNonMarginableBuyingPower retrieves the non-marginable buying power in the account
func (c *Client) GetNonMarginableBuyingPower(ctx context.Context) (float64, error) {
	account, err := c.tradingClient.GetAccount()
	if err != nil {
		return 0, fmt.Errorf("failed to get account: %w", err)
	}

	log.Printf("Cash balance: %s", account.Cash)
	log.Printf("Non-marginable buying power: %s", account.NonMarginBuyingPower)

	return account.NonMarginBuyingPower.InexactFloat64(), nil
}

// GetOptionsPositions retrieves all option positions for a specific underlying ticker
func (c *Client) GetOptionsPositions(ctx context.Context, underlyingTicker string) ([]alpaca.Position, error) {
	if underlyingTicker == "" {
		return nil, fmt.Errorf("underlying ticker cannot be empty")
	}

	// Get all positions
	allPositions, err := c.GetAllPositions(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all positions: %w", err)
	}

	var optionPositions []alpaca.Position

	// Filter positions by underlying ticker
	for _, position := range allPositions {
		// Parse the position symbol to check if it's an option
		parsedOption, err := c.ParseOptionTicker(position.Symbol)
		if err != nil {
			// If parsing fails, it's not an option, so skip it
			continue
		}

		// Check if the underlying matches the requested ticker
		if parsedOption.Underlying == underlyingTicker {
			optionPositions = append(optionPositions, position)
		}
	}

	return optionPositions, nil
}
