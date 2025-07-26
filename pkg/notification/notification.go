package notification

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

// Client wraps the notification webhook client and supports multiple methods
// Supported methods: "generic", "discord"
type Client struct {
	noisyWebhookURL  string
	normalWebhookURL string
	method           string // "generic" or "discord"
}

// NewClient creates a new notification client using the NOTIFY_WEBHOOK_URL and NOTIFY_METHOD environment variables
func NewClient() (*Client, error) {
	noisyWebhookURL := os.Getenv("NOTIFY_NOISY_WEBHOOK_URL")
	normalWebhookURL := os.Getenv("NOTIFY_NORMAL_WEBHOOK_URL")
	method := os.Getenv("NOTIFY_METHOD")
	if method == "" {
		method = "generic"
	}
	return &Client{noisyWebhookURL: noisyWebhookURL, normalWebhookURL: normalWebhookURL, method: method}, nil
}

type payload struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

func (c *Client) sendNotification(webhookURL, notificationType, message string) error {
	if webhookURL == "" {
		// No-op if webhookURL is not set
		return nil
	}

	var b []byte
	var err error
	var contentType string

	switch c.method {
	case "discord":
		// Discord expects: {"content": "<emoji + type + message> @everyone", "allowed_mentions":{"parse":["everyone"]}}
		fullMessage := notificationType + ": " + message + " @everyone"
		discordPayload := map[string]interface{}{
			"content":          fullMessage,
			"allowed_mentions": map[string]interface{}{"parse": []string{"everyone"}},
		}
		b, err = json.Marshal(discordPayload)
		contentType = "application/json"
	case "generic":
		fallthrough
	default:
		p := payload{
			Type:    notificationType,
			Message: message,
		}
		b, err = json.Marshal(p)
		contentType = "application/json"
	}

	if err != nil {
		return err
	}
	resp, err := http.Post(webhookURL, contentType, bytes.NewBuffer(b))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("notification failed with status: %s", resp.Status)
	}
	return nil
}

func (c *Client) OrderPlaced(message string) error {
	_ = c.sendNotification(c.normalWebhookURL, "✅ Order Placed", message)
	return nil
}

func (c *Client) Failure(message string) error {
	_ = c.sendNotification(c.normalWebhookURL, "❌ Error occurred", message)
	return fmt.Errorf("%s", message)
}

func (c *Client) ActionNeeded(message string, err error) error {
	_ = c.sendNotification(c.normalWebhookURL, "⚠️ Action needed", message)
	return err
}

func (c *Client) MaxActiveOptions(message string) error {
	_ = c.sendNotification(c.normalWebhookURL, "⏩ Skipping", message)
	return nil
}

func (c *Client) NoGapDown(message string) error {
	_ = c.sendNotification(c.noisyWebhookURL, "🚫 No gap down", message)
	return nil
}

func (c *Client) MarketClosed() error {
	msg := fmt.Sprintf("The market is closed on %s", time.Now().Format("January 2, 2006"))
	_ = c.sendNotification(c.noisyWebhookURL, "🚫 Market closed", msg)
	return nil
}
