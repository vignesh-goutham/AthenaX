package alpaca

import (
	"fmt"
	"os"

	"github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
	"github.com/alpacahq/alpaca-trade-api-go/v3/marketdata"
)

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
