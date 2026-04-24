package email

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Config struct {
	BaseURL     string
	FromAddress string
	FromName    string
}

type EmailClient interface {
	SendEmail(ctx context.Context, to, subject, body string) error
}

type emailClient struct {
	baseURL    string
	fromAddr   string
	fromName   string
	httpClient *http.Client
}

func NewEmailClient(cfg Config) EmailClient {
	return &emailClient{
		baseURL:  cfg.BaseURL,
		fromAddr: cfg.FromAddress,
		fromName: cfg.FromName,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type sendEmailRequest struct {
	To          string `json:"to"`
	From        string `json:"from"`
	FromName    string `json:"from_name"`
	Subject     string `json:"subject"`
	Body        string `json:"body"`
	ContentType string `json:"content_type"`
}

func (c *emailClient) SendEmail(ctx context.Context, to, subject, body string) error {
	reqBody := sendEmailRequest{
		To:          to,
		From:        c.fromAddr,
		FromName:    c.fromName,
		Subject:     subject,
		Body:        body,
		ContentType: "text/html",
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("marshalling email request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v1/send", bytes.NewReader(jsonBody))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("sending email: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("email service returned status %d", resp.StatusCode)
	}

	return nil
}
