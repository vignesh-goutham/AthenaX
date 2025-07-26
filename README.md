# AthenaX

A precision trading bot inspired by Athena, the goddess of strategic warfare and wisdom. AthenaX executes defined trading strategies with the precision and intelligence of its namesake.

## Quick Start

### Build
```bash
make build
```

### Run a Strategy
```bash
./_bin/athenax run-strategy --name two-percent-down
```

### Available Strategies
- **two-percent-down**: Executes a 2% gap down strategy

## Configuration

Set your Alpaca API credentials:
```bash
export ALPACA_API_KEY="your_api_key"
export ALPACA_SECRET_KEY="your_secret_key"
```