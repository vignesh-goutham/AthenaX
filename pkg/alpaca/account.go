package alpaca

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
	"github.com/alpacahq/alpaca-trade-api-go/v3/marketdata"
	"github.com/shopspring/decimal"
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

// PlaceOptionLimitOrderWithTakeProfit places a bracket order for an option with entry at 99% of ask price and take profit
// Since options don't support fractional shares, it calculates the appropriate quantity
// takeProfitPercentage is a percentage (e.g., 20.0 means 20% profit)
func (m *Client) PlaceOptionLimitOrderWithTakeProfit(ctx context.Context, investmentSize float64, optionSymbol string, optionQuote *marketdata.OptionQuote, takeProfitPercentage float64) (*alpaca.Order, error) {
	if optionSymbol == "" {
		return nil, fmt.Errorf("option symbol cannot be empty")
	}

	if optionQuote == nil {
		return nil, fmt.Errorf("option quote cannot be nil")
	}

	if investmentSize <= 0 {
		return nil, fmt.Errorf("investment size must be greater than 0")
	}

	if takeProfitPercentage <= 0 {
		return nil, fmt.Errorf("take profit percentage must be greater than 0")
	}

	// Calculate limit price as 99% of ask price
	if optionQuote.BidPrice <= 0 || optionQuote.AskPrice <= 0 {
		return nil, fmt.Errorf("invalid bid/ask prices: bid=%.2f, ask=%.2f", optionQuote.BidPrice, optionQuote.AskPrice)
	}

	limitPrice := optionQuote.AskPrice * 0.99
	// Round to 2 decimal places for Alpaca API compliance
	limitPrice = float64(int(limitPrice*100)) / 100

	// Calculate take profit price as a percentage of the limit price
	takeProfitPrice := limitPrice * (1 + takeProfitPercentage/100)
	// Round to 2 decimal places for Alpaca API compliance
	takeProfitPrice = float64(int(takeProfitPrice*100)) / 100

	// Calculate quantity (options are typically sold in contracts of 100 shares)
	// Each option contract represents 100 shares of the underlying
	quantity := int(investmentSize / (optionQuote.AskPrice * 100))

	if quantity <= 0 {
		return nil, fmt.Errorf("calculated quantity is 0 or negative: investment=%.2f, askPrice=%.2f, quantity=%d",
			investmentSize, optionQuote.AskPrice, quantity)
	}

	// Calculate actual order value
	actualOrderValue := float64(quantity) * optionQuote.AskPrice * 100

	log.Printf("Placing bracket order: symbol=%s, quantity=%d contracts, limitPrice=%.2f, orderValue=%.2f, takeProfit=%.1f%% (price=%.2f)",
		optionSymbol, quantity, limitPrice, actualOrderValue, takeProfitPercentage, takeProfitPrice)

	// Place the bracket order
	qty := decimal.NewFromFloat(float64(quantity))
	limitPriceDecimal := decimal.NewFromFloat(limitPrice)
	takeProfitPriceDecimal := decimal.NewFromFloat(takeProfitPrice)

	order, err := m.tradingClient.PlaceOrder(alpaca.PlaceOrderRequest{
		Symbol:      optionSymbol,
		Qty:         &qty,
		Side:        alpaca.Buy,
		Type:        alpaca.Limit,
		TimeInForce: alpaca.Day,
		LimitPrice:  &limitPriceDecimal,
		TakeProfit:  &alpaca.TakeProfit{LimitPrice: &takeProfitPriceDecimal},
	})

	if err != nil {
		return nil, fmt.Errorf("failed to place bracket order: %w", err)
	}

	log.Printf("Bracket order placed successfully: ID=%s, Status=%s", order.ID, order.Status)
	return order, nil
}

// getLastTradingDay uses Alpaca calendar API to get the actual last trading day
func (m *Client) getLastTradingDay(ctx context.Context) (time.Time, error) {
	// Get calendar for the last 10 days to find the most recent trading day
	end := time.Now()
	start := end.AddDate(0, 0, -10)

	calendar, err := m.tradingClient.GetCalendar(alpaca.GetCalendarRequest{
		Start: start,
		End:   end,
	})
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to get calendar: %w", err)
	}

	if len(calendar) == 0 {
		return time.Time{}, fmt.Errorf("no calendar data available")
	}

	// Get today's date (without time component) for comparison
	today := time.Now().Truncate(24 * time.Hour)
	var lastTradingDay time.Time

	// Find the most recent trading day that's strictly before today
	for i := len(calendar) - 1; i >= 0; i-- {
		day := calendar[i]
		// Parse the date string (format: "2024-01-15")
		date, err := time.Parse("2006-01-02", day.Date)
		if err != nil {
			continue // Skip invalid dates
		}

		// Compare dates without time component
		if date.Before(today) {
			lastTradingDay = date
			break
		}
	}

	if lastTradingDay.IsZero() {
		return time.Time{}, fmt.Errorf("no previous trading day found")
	}

	return lastTradingDay, nil
}

// IsMarketOpen checks if the market is currently open
func (m *Client) IsMarketOpen(ctx context.Context) (bool, error) {
	clock, err := m.tradingClient.GetClock()
	if err != nil {
		return false, fmt.Errorf("failed to get market clock: %w", err)
	}

	return clock.IsOpen, nil
}
