package alpaca

import (
	"context"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"time"

	"cloud.google.com/go/civil"
	"github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
	"github.com/alpacahq/alpaca-trade-api-go/v3/marketdata"
	"github.com/shopspring/decimal"
)

// Option represents a parsed option ticker
type Option struct {
	Underlying string    // Underlying ticker (e.g., "QQQ")
	Expiry     time.Time // Expiration date
	Type       string    // "C" for call, "P" for put
	Strike     float64   // Strike price
	Ticker     string    // Full option symbol
}

// ParseOptionTicker parses an option ticker symbol and returns structured data
func (m *Client) ParseOptionTicker(symbol string) (*Option, error) {
	if symbol == "" {
		return nil, fmt.Errorf("symbol cannot be empty")
	}

	// Option symbols follow the OSI format: TICKERYYMMDDC/PSTRIKE
	// Example: QQQ240119C00420000 (QQQ, 2024-01-19, Call, $420.00)

	// Find the strike price (last 8 digits)
	if len(symbol) < 8 {
		return nil, fmt.Errorf("invalid option symbol format: too short")
	}

	strikeStr := symbol[len(symbol)-8:]
	strikeFloat, err := strconv.ParseFloat(strikeStr, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid strike price: %w", err)
	}

	// Convert strike from integer format (e.g., 00420000) to decimal (420.00)
	strike := strikeFloat / 1000

	// Remove strike from symbol to get ticker + date + type
	baseSymbol := symbol[:len(symbol)-8]

	// Find the option type (C or P) - it's the character before the strike
	if len(baseSymbol) < 1 {
		return nil, fmt.Errorf("invalid option symbol format: missing option type")
	}

	optionTypeChar := baseSymbol[len(baseSymbol)-1]
	var optionType string
	if optionTypeChar == 'C' {
		optionType = "C"
	} else if optionTypeChar == 'P' {
		optionType = "P"
	} else {
		return nil, fmt.Errorf("invalid option type: expected C or P, got %c", optionTypeChar)
	}

	// Remove option type to get ticker + date
	datePart := baseSymbol[:len(baseSymbol)-1]

	// Extract date (YYMMDD format)
	if len(datePart) < 6 {
		return nil, fmt.Errorf("invalid option symbol format: missing date")
	}

	dateStr := datePart[len(datePart)-6:]
	year := "20" + dateStr[:2] // Assume 20xx years
	month := dateStr[2:4]
	day := dateStr[4:6]

	// Parse the date
	dateLayout := "2006-01-02"
	dateStrFull := fmt.Sprintf("%s-%s-%s", year, month, day)
	expiry, err := time.Parse(dateLayout, dateStrFull)
	if err != nil {
		return nil, fmt.Errorf("invalid expiration date: %w", err)
	}

	// Extract ticker (everything before the date)
	underlying := datePart[:len(datePart)-6]

	return &Option{
		Underlying: underlying,
		Expiry:     expiry,
		Type:       optionType,
		Strike:     strike,
		Ticker:     symbol,
	}, nil
}

// Client wraps the Alpaca market data client
type Client struct {
	marketDataClient *marketdata.Client
	tradingClient    *alpaca.Client
}

// NewClient creates a new client using environment variables
func NewClient() (*Client, error) {
	apiKey := os.Getenv("ALPACA_API_KEY")
	secretKey := os.Getenv("ALPACA_SECRET_KEY")

	if apiKey == "" || secretKey == "" {
		return nil, fmt.Errorf("ALPACA_API_KEY and ALPACA_SECRET_KEY environment variables must be set")
	}

	marketDataClient := marketdata.NewClient(marketdata.ClientOpts{
		APIKey:    apiKey,
		APISecret: secretKey,
	})

	tradingClient := alpaca.NewClient(alpaca.ClientOpts{
		APIKey:    apiKey,
		APISecret: secretKey,
		BaseURL:   "https://paper-api.alpaca.markets",
	})

	return &Client{
		marketDataClient: marketDataClient,
		tradingClient:    tradingClient,
	}, nil
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

// IsMarketOpen checks if the market is currently open
func (m *Client) IsMarketOpen(ctx context.Context) (bool, error) {
	clock, err := m.tradingClient.GetClock()
	if err != nil {
		return false, fmt.Errorf("failed to get market clock: %w", err)
	}

	return clock.IsOpen, nil
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
