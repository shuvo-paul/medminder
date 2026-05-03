package email

import (
	"context"
	"log/slog"
	"sync"
)

// EmailJob represents an email to be sent.
type EmailJob struct {
	To      string
	Subject string
	Body    string
}

// emailQueue manages the email sending queue.
type emailQueue struct {
	jobs    chan EmailJob
	client  EmailClient
	logger  *slog.Logger
	workers int
	wg      sync.WaitGroup
}

var (
	globalQueue *emailQueue
	queueOnce  sync.Once
)

// QueueEmail adds an email job to the queue.
func QueueEmail(ctx context.Context, to, subject, body string) {
	if globalQueue == nil {
		return
	}
	select {
	case <-ctx.Done():
	default:
		globalQueue.jobs <- EmailJob{To: to, Subject: subject, Body: body}
	}
}

// StartEmailQueue initializes the global email queue with worker pool.
func StartEmailQueue(client EmailClient, workers int, logger *slog.Logger) {
	queueOnce.Do(func() {
		globalQueue = &emailQueue{
			jobs:    make(chan EmailJob, 100),
			client:  client,
			logger:  logger,
			workers: workers,
		}
		for i := 0; i < workers; i++ {
			globalQueue.wg.Add(1)
			go globalQueue.worker()
		}
	})
}

// StopEmailQueue gracefully shuts down the email queue.
func StopEmailQueue() {
	if globalQueue == nil {
		return
	}
	close(globalQueue.jobs)
	globalQueue.wg.Wait()
}

func (q *emailQueue) worker() {
	defer q.wg.Done()
	for job := range q.jobs {
		ctx := context.Background()
		if err := q.client.SendEmail(ctx, job.To, job.Subject, job.Body); err != nil {
			q.logger.Error("failed to send email", "to", job.To, "subject", job.Subject, "error", err)
		}
	}
}