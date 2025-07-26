# AthenaX

A precision trading bot inspired by Athena, the goddess of strategic warfare and wisdom. AthenaX executes defined trading strategies with the precision and intelligence of its namesake.

## Quick Start

### Build
```bash
make build
```

### Build for AWS Lambda
```bash
make build-lambda
make package-lambda
```

### Run a Strategy locally
```bash
./_bin/athenax run-strategy --name two-percent-down
```

### Available Strategies
- **two-percent-down**: When QQQ gaps down 2% or more at runtime, automatically places a bracket order to buy a LEAP call option with delta >= 0.60, setting a take profit target at 50% gain.

## Configuration

### Environment Variables

Set the following environment variables before running AthenaX:

#### Alpaca Trading API
```bash
export ALPACA_API_KEY="your_api_key"
export ALPACA_SECRET_KEY="your_secret_key"
```

#### Notification System
```bash
# Webhook URLs for notifications
export NOTIFY_NOISY_WEBHOOK_URL="https://your-webhook-url.com/noisy"
export NOTIFY_NORMAL_WEBHOOK_URL="https://your-webhook-url.com/normal"

# Notification method: "generic" or "discord" (default: "generic")
export NOTIFY_METHOD="discord"
```

#### Strategy Configuration
```bash
# Maximum number of active options (default: 5)
export MAX_ACTIVE_OPTIONS="5"
```

### Alpaca Setup

AthenaX uses Alpaca's paper trading environment by default. To get started:

1. **Create an Alpaca Account**: Sign up at [alpaca.markets](https://alpaca.markets)
2. **Get API Credentials**: Navigate to your Alpaca dashboard and generate API keys
3. **Set Environment Variables**: Export your API key and secret key as shown above
4. **Fund Your Account**: Add funds to your paper trading account for testing

**Note**: The bot currently uses Alpaca's paper trading API (`https://paper-api.alpaca.markets`). For live trading, you'll need to modify the base URL in the code.

**Important**: This bot requires an Alpaca Pro subscription as it uses the SIP (Securities Information Processor) feed to get NBBO (National Best Bid and Offer) and live quotes for accurate market data.

### Notification System

AthenaX supports real-time notifications for trading events through webhooks:

#### Supported Notification Methods
- **Generic**: Standard JSON webhook format
- **Discord**: Discord-specific format with mentions

#### Notification Types
- ✅ **Order Placed**: Successful order execution
- ❌ **Error occurred**: Trading or system errors
- ⚠️ **Action needed**: Requires manual intervention
- ⏩ **Skipping**: Strategy skipped (e.g., max options reached)
- 🚫 **No gap down**: No significant market movement detected
- 🚫 **Market closed**: Market is currently closed

#### Webhook Configuration
- **Noisy Webhook**: Used for frequent, less critical notifications (e.g., "no gap down", "market closed")
- **Normal Webhook**: Used for important trading events (e.g., orders placed, errors)

#### Discord Integration
When using Discord notifications (`NOTIFY_METHOD="discord"`):
- Messages include `@everyone` mentions
- Emojis are automatically added based on notification type
- Messages are formatted for Discord's webhook API

#### Example Discord Webhook Setup
1. Create a Discord server channel
2. Go to Channel Settings → Integrations → Webhooks
3. Create a new webhook and copy the URL
4. Set `NOTIFY_NORMAL_WEBHOOK_URL` and/or `NOTIFY_NOISY_WEBHOOK_URL` to your Discord webhook URL
5. Set `NOTIFY_METHOD="discord"`