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
- **two-percent-down**: When QQQ gaps down 2% or more at market open, automatically places a bracket order to buy a LEAP call option with delta >= 0.60, setting a take profit target at 50% gain.

## Configuration

Set your Alpaca API credentials:
```bash
export ALPACA_API_KEY="your_api_key"
export ALPACA_SECRET_KEY="your_secret_key"
```