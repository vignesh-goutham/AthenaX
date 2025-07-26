package alpaca

import (
	"fmt"
	"strconv"
	"time"
)

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
