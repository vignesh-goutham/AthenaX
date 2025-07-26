package strategies

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/vignesh-goutham/AthenaX/pkg/alpaca"
)

const ticker = "QQQ"

type TwoPercentDown struct {
	broker           *alpaca.Client
	maxActiveOptions int
}

// NewTwoPercentDown creates a new TwoPercentDown strategy instance
func NewTwoPercentDown(broker *alpaca.Client) *TwoPercentDown {
	// Get max active options from environment variable, default to 5
	maxActiveOptions := 5
	if envMax := os.Getenv("MAX_ACTIVE_OPTIONS"); envMax != "" {
		if parsed, err := strconv.Atoi(envMax); err == nil && parsed > 0 {
			maxActiveOptions = parsed
		}
	}

	return &TwoPercentDown{
		broker:           broker,
		maxActiveOptions: maxActiveOptions,
	}
}

func (s *TwoPercentDown) Run(ctx context.Context) error {
	// Step 1: Get yesterday's close of ticker
	yesterdayClose, err := s.broker.GetLastTradingDayClose(ctx, ticker)
	if err != nil {
		return fmt.Errorf("failed to get yesterday's close for %s: %w", ticker, err)
	}

	// Step 2: Get latest quote now
	currentPrice, err := s.broker.GetLatestQuote(ctx, ticker)
	if err != nil {
		return fmt.Errorf("failed to get latest quote for %s: %w", ticker, err)
	}

	// Step 3: Calculate gap down if any
	changePercent := ((currentPrice - yesterdayClose) / yesterdayClose) * 100

	// Step 4: If it's 2% or more gap down, print it's a gapdown
	if changePercent <= -2.0 {
		log.Printf("GAP DOWN DETECTED: %s is down %.2f%% from yesterday's close (Current: $%.2f, Yesterday: $%.2f)",
			ticker, -changePercent, currentPrice, yesterdayClose)

		// Step 5: Get the lowest strike call LEAPS option with delta >= 0.6
		optionSymbol, optionSnapshot, err := s.broker.GetCallLeapsByDelta(ctx, ticker, 0.60)
		if err != nil {
			return fmt.Errorf("failed to get call LEAPS option for %s: %w", ticker, err)
		}
		fmt.Printf("Found option symbol: %s\n", optionSymbol)
		fmt.Printf("Found option snapshot: %+v\n", optionSnapshot)

		// Calculate investment size for this option
		investmentSize, err := s.calculateInvestmentSize(ctx)
		if err != nil {
			return fmt.Errorf("failed to calculate investment size: %w", err)
		}

		log.Printf("Will invest $%.2f in option %s", investmentSize, optionSymbol)

		// Place the order
		_, err = s.broker.PlaceOptionLimitOrderWithTakeProfit(ctx, investmentSize, optionSymbol, optionSnapshot.LatestQuote, 50.0)
		if err != nil {
			return fmt.Errorf("failed to place order: %w", err)
		}

	} else {
		log.Printf("No significant gap down: %s is %+.2f%% from yesterday's close (Current: $%.2f, Yesterday: $%.2f)",
			ticker, changePercent, currentPrice, yesterdayClose)
	}

	return nil
}

// calculateInvestmentSize determines the investment size per option based on remaining spots and buying power
func (s *TwoPercentDown) calculateInvestmentSize(ctx context.Context) (float64, error) {
	// Get all QQQ option positions
	openOptions, err := s.broker.GetOptionsPositions(ctx, ticker)
	if err != nil {
		return 0, fmt.Errorf("failed to get QQQ option positions: %w", err)
	}

	// Calculate remaining active option spots
	remainingSpots := s.maxActiveOptions - len(openOptions)
	if remainingSpots <= 0 {
		return 0, fmt.Errorf("no remaining active option spots available")
	}

	// Get non-marginable buying power
	buyingPower, err := s.broker.GetNonMarginableBuyingPower(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get non-marginable buying power: %w", err)
	}

	// Calculate investment size per option
	investmentSize := buyingPower / float64(remainingSpots)

	log.Printf("Investment calculation: Buying power $%.2f / %d remaining spots = $%.2f per trade",
		buyingPower, remainingSpots, investmentSize)

	return investmentSize, nil
}
