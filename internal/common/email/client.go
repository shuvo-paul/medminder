package email

import (
	"context"
	"fmt"
	"time"

	"github.com/shuvo-paul/medminder/internal/common/config"
	"github.com/wneessen/go-mail"
)

type EmailClient interface {
	SendEmail(ctx context.Context, to, subject, body string) error
}

type emailClient struct {
	host          string
	port          int
	username      string
	password      string
	fromAddr      string
	fromName      string
	clientTimeout time.Duration
}

func NewEmailClient(cfg config.EmailConfig) EmailClient {
	return &emailClient{
		host:          cfg.SMTPHost,
		port:          cfg.SMTPPort,
		username:      cfg.SMTPUsername,
		password:      cfg.SMTPPassword,
		fromAddr:      cfg.FromAddress,
		fromName:      cfg.FromName,
		clientTimeout: 10 * time.Second,
	}
}

func (c *emailClient) SendEmail(ctx context.Context, to, subject, body string) error {
	client, err := mail.NewClient(c.host,
		mail.WithPort(c.port),
		mail.WithUsername(c.username),
		mail.WithPassword(c.password),
		mail.WithTLSPolicy(mail.TLSOpportunistic),
		mail.WithSMTPAuth(mail.SMTPAuthPlain),
		mail.WithTimeout(c.clientTimeout),
	)
	if err != nil {
		return fmt.Errorf("creating mail client: %w", err)
	}

	msg := mail.NewMsg()
	if err := msg.FromFormat(c.fromName, c.fromAddr); err != nil {
		return fmt.Errorf("setting from address: %w", err)
	}
	if err := msg.To(to); err != nil {
		return fmt.Errorf("setting to address: %w", err)
	}
	msg.Subject(subject)
	msg.SetBodyString(mail.TypeTextHTML, body)

	if err := client.DialAndSendWithContext(ctx, msg); err != nil {
		return fmt.Errorf("sending email: %w", err)
	}

	return nil
}
