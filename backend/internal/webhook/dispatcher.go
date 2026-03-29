// Package webhook dispatches digest summaries to user-configured webhook URLs.
package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/hadfielj/taran/backend/internal/domain"
)

// Payload is the JSON body sent to webhook endpoints.
type Payload struct {
	Event       string    `json:"event"`
	DigestID    string    `json:"digestId"`
	Title       string    `json:"title"`
	Summary     string    `json:"summary"`
	Highlights  []string  `json:"highlights,omitempty"`
	TopTopics   []string  `json:"topTopics,omitempty"`
	EmailCount  int       `json:"emailCount"`
	PeriodStart time.Time `json:"periodStart"`
	PeriodEnd   time.Time `json:"periodEnd"`
	ViewURL     string    `json:"viewUrl,omitempty"`
}

// SendDigest posts a digest summary to the given webhook URL.
func SendDigest(ctx context.Context, webhookURL string, digest *domain.Digest, viewURL string) error {
	payload := Payload{
		Event:       "digest.created",
		DigestID:    digest.ID,
		Title:       digest.Title,
		Summary:     digest.Summary,
		Highlights:  digest.Highlights,
		TopTopics:   digest.TopTopics,
		EmailCount:  digest.EmailCount,
		PeriodStart: digest.PeriodStart,
		PeriodEnd:   digest.PeriodEnd,
		ViewURL:     viewURL,
	}

	if err := ValidateExternalURL(webhookURL); err != nil {
		return fmt.Errorf("webhook URL blocked: %w", err)
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal webhook payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, webhookURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create webhook request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "MailBrief/1.0")

	resp, err := SafeHTTPClient().Do(req)
	if err != nil {
		return fmt.Errorf("webhook request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}

	slog.Info("webhook delivered", "url", webhookURL, "status", resp.StatusCode, "digestID", digest.ID)
	return nil
}
