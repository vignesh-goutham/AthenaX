package alpaca

import (
	"context"
	"fmt"
	"log"
	"sort"
	"time"

	"cloud.google.com/go/civil"
	"github.com/alpacahq/alpaca-trade-api-go/v3/marketdata"
)

// GetCallLeapsByDelta finds the lowest strike call LEAPS option with delta >= 60
// LEAPS are options with expiration > 11 months from current date
func (m *Client) GetCallLeapsByDelta(ctx context.Context, underlyingTicker string, minDelta float64) (string, *marketdata.OptionSnapshot, error) {
	if underlyingTicker == "" {
		return "", nil, fmt.Errorf("underlying ticker cannot be empty")
	}

	if minDelta <= 0 {
		return "", nil, fmt.Errorf("minimum delta must be greater than 0")
	}

	// Calculate expiration date threshold (11 months from now)
	elevenMonthsFromNow := time.Now().AddDate(0, 11, 0)
	expirationDateGte := civil.DateOf(elevenMonthsFromNow)

	// Get option chain for the underlying symbol
	optionChain, err := m.marketDataClient.GetOptionChain(underlyingTicker, marketdata.GetOptionChainRequest{
		Type:              marketdata.Call,
		ExpirationDateGte: expirationDateGte,
		Feed:              marketdata.OPRA,
		TotalLimit:        1000, // Get a reasonable number of options
	})
	if err != nil {
		return "", nil, fmt.Errorf("failed to get option chain for %s: %w", underlyingTicker, err)
	}

	if len(optionChain) == 0 {
		return "", nil, fmt.Errorf("no call LEAPS options found for %s", underlyingTicker)
	}

	// Filter options with delta >= minDelta and sort by strike price
	var validOptions []struct {
		symbol   string
		snapshot *marketdata.OptionSnapshot
		delta    float64
		expiry   time.Time
	}

	for symbol, snapshot := range optionChain {
		// Check if Greeks data is available
		if snapshot.Greeks == nil {
			continue
		}

		delta := snapshot.Greeks.Delta
		// For call options, we want positive delta
		if delta >= minDelta {
			// Parse the option symbol to get expiry date
			option, err := m.ParseOptionTicker(symbol)
			if err != nil {
				log.Printf("Failed to parse option ticker %s: %v", symbol, err)
				continue
			}

			validOptions = append(validOptions, struct {
				symbol   string
				snapshot *marketdata.OptionSnapshot
				delta    float64
				expiry   time.Time
			}{
				symbol:   symbol,
				snapshot: &snapshot,
				delta:    delta,
				expiry:   option.Expiry,
			})
		}
	}

	if len(validOptions) == 0 {
		return "", nil, fmt.Errorf("no call LEAPS options found for %s with delta >= %.2f", underlyingTicker, minDelta)
	}

	// Sort by expiry date (earliest first)
	sort.Slice(validOptions, func(i, j int) bool {
		return validOptions[i].expiry.Before(validOptions[j].expiry)
	})

	// Get the earliest expiry date
	earliestExpiry := validOptions[0].expiry

	// Filter to only options with the earliest expiry
	var earliestExpiryOptions []struct {
		symbol   string
		snapshot *marketdata.OptionSnapshot
		delta    float64
		expiry   time.Time
	}

	for _, option := range validOptions {
		if option.expiry.Equal(earliestExpiry) {
			earliestExpiryOptions = append(earliestExpiryOptions, option)
		}
	}

	// Find the option with minimum delta from the earliest expiry
	minDeltaOption := earliestExpiryOptions[0]
	for _, option := range earliestExpiryOptions {
		if option.delta < minDeltaOption.delta {
			minDeltaOption = option
		}
	}

	// Return the option with the minimum delta from the earliest expiry
	return minDeltaOption.symbol, minDeltaOption.snapshot, nil
}

// GetLatestBar retrieves the latest bar for a symbol
func (m *Client) GetLatestBar(ctx context.Context, symbol string) (*marketdata.Bar, error) {
	if symbol == "" {
		return nil, fmt.Errorf("symbol cannot be empty")
	}

	bar, err := m.marketDataClient.GetLatestBar(symbol, marketdata.GetLatestBarRequest{
		Feed: marketdata.SIP,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get latest bar for %s: %w", symbol, err)
	}

	log.Printf("Latest bar data: %+v", bar)
	return bar, nil
}

// GetLatestQuote retrieves the latest quote for a symbol and returns the ask price
func (m *Client) GetLatestQuote(ctx context.Context, symbol string) (float64, error) {
	if symbol == "" {
		return 0, fmt.Errorf("symbol cannot be empty")
	}

	quote, err := m.marketDataClient.GetLatestQuote(symbol, marketdata.GetLatestQuoteRequest{
		Feed: marketdata.SIP,
	})
	if err != nil {
		return 0, fmt.Errorf("failed to get latest quote for %s: %w", symbol, err)
	}

	log.Printf("Latest quote data: %+v", quote)

	// Return ask price for buying scenarios
	return quote.AskPrice, nil
}

// GetLatestBarMidPrice calculates and returns the mid price from the latest bar
func (m *Client) GetLatestBarMidPrice(ctx context.Context, symbol string) (float64, error) {
	bar, err := m.GetLatestBar(ctx, symbol)
	if err != nil {
		return 0, err
	}

	// Calculate mid price as (high + low) / 2
	midPrice := (bar.High + bar.Low) / 2
	return midPrice, nil
}

// GetLastTradingDayClose retrieves the closing price for the last trading day
func (m *Client) GetLastTradingDayClose(ctx context.Context, symbol string) (float64, error) {
	if symbol == "" {
		return 0, fmt.Errorf("symbol cannot be empty")
	}

	// Get the last trading day using Alpaca calendar API
	lastTradingDay, err := m.getLastTradingDay(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get last trading day: %w", err)
	}

	log.Printf("Last trading day: %s\n", lastTradingDay.Format("2006-01-02"))

	// Get daily bars for the symbol
	bars, err := m.marketDataClient.GetBars(symbol, marketdata.GetBarsRequest{
		TimeFrame:  marketdata.OneDay,
		Start:      lastTradingDay,
		End:        lastTradingDay.Add(24 * time.Hour),
		Feed:       marketdata.SIP,
		TotalLimit: 1,
	})
	if err != nil {
		return 0, fmt.Errorf("failed to get bars for %s: %w", symbol, err)
	}

	if len(bars) == 0 {
		return 0, fmt.Errorf("no data found for %s on %s", symbol, lastTradingDay.Format("2006-01-02"))
	}

	log.Printf("Last trading day bars data: %+v", bars[0])

	return bars[0].Close, nil
}
