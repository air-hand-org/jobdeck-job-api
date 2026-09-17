// Package eventing contains job API event publication and outbox dispatching.
package eventing

import (
	"context"

	"github.com/air-hand-org/jobdeck-job-api/internal/domain"
)

// ExecutionPublisher publishes execution-related events.
//
//go:generate moq -out gen_execution_publisher_moq_test.go . ExecutionPublisher
type ExecutionPublisher interface {
	Publish(ctx context.Context, executionID domain.ExecutionId) error
}
