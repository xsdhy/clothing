package llm

import (
	"clothing/internal/entity"
	"context"
	"errors"
	"time"

	"github.com/sirupsen/logrus"
)

// TaskStatus represents the status of an async generation task.
type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusRunning   TaskStatus = "running"
	TaskStatusSucceeded TaskStatus = "succeeded"
	TaskStatusFailed    TaskStatus = "failed"
	TaskStatusCancelled TaskStatus = "cancelled"
)

// AsyncTask represents an asynchronous generation task.
type AsyncTask struct {
	ID         string
	ProviderID string
	ModelID    string
	Status     TaskStatus
	Progress   float64
	Result     *entity.GenerateContentResponse
	Error      error
	CreatedAt  time.Time
}

// PollConfig contains configuration for polling async tasks.
type PollConfig struct {
	Interval    time.Duration
	MaxAttempts int
	Backoff     bool
	BackoffMax  time.Duration
}

// DefaultPollConfig provides default polling configuration.
var DefaultPollConfig = PollConfig{
	Interval:    5 * time.Second,
	MaxAttempts: 120, // 10 minutes with 5s interval
	Backoff:     false,
	BackoffMax:  30 * time.Second,
}

// FalAIPollConfig provides polling configuration optimized for FalAI.
var FalAIPollConfig = PollConfig{
	Interval:    2 * time.Second,
	MaxAttempts: 60,
	Backoff:     false,
}

// VolcenginePollConfig provides polling configuration for Volcengine.
var VolcenginePollConfig = PollConfig{
	Interval:    5 * time.Second,
	MaxAttempts: 120,
	Backoff:     false,
}

// TaskPoller defines the interface for polling task status.
type TaskPoller interface {
	// Poll checks the current status of a task.
	Poll(ctx context.Context, taskID string) (*AsyncTask, error)
}

// WaitForTask polls a task until completion or timeout.
func WaitForTask(ctx context.Context, poller TaskPoller, taskID string, config PollConfig) (*entity.GenerateContentResponse, error) {
	if taskID == "" {
		return nil, errors.New("task ID is required")
	}

	interval := config.Interval
	if interval <= 0 {
		interval = DefaultPollConfig.Interval
	}

	maxAttempts := config.MaxAttempts
	if maxAttempts <= 0 {
		maxAttempts = DefaultPollConfig.MaxAttempts
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	attempts := 0

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()

		case <-ticker.C:
			attempts++

			task, err := poller.Poll(ctx, taskID)
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"task_id": taskID,
					"attempt": attempts,
					"error":   err,
				}).Warn("task_manager: poll error")
				return nil, err
			}

			logrus.WithFields(logrus.Fields{
				"task_id":  taskID,
				"status":   task.Status,
				"progress": task.Progress,
				"attempt":  attempts,
			}).Debug("task_manager: poll status")

			switch task.Status {
			case TaskStatusSucceeded:
				return task.Result, nil

			case TaskStatusFailed:
				if task.Error != nil {
					return nil, task.Error
				}
				return nil, errors.New("task failed without error message")

			case TaskStatusCancelled:
				return nil, errors.New("task was cancelled")

			case TaskStatusPending, TaskStatusRunning:
				if attempts >= maxAttempts {
					return nil, errors.New("polling exceeded maximum attempts")
				}

				// Apply backoff if enabled
				if config.Backoff {
					newInterval := interval * 2
					if config.BackoffMax > 0 && newInterval > config.BackoffMax {
						newInterval = config.BackoffMax
					}
					if newInterval != interval {
						ticker.Reset(newInterval)
						interval = newInterval
					}
				}

			default:
				// Unknown status, continue polling
				if attempts >= maxAttempts {
					return nil, errors.New("polling exceeded maximum attempts with unknown status")
				}
			}
		}
	}
}

// MapTaskStatus maps provider-specific status strings to TaskStatus.
// This provides a unified way to handle status from different providers.
func MapTaskStatus(status string) TaskStatus {
	// Normalize to lowercase for comparison
	normalized := toLowerASCII(status)

	switch normalized {
	case "pending", "queued", "in_queue", "created":
		return TaskStatusPending
	case "running", "processing", "in_progress", "started":
		return TaskStatusRunning
	case "succeeded", "success", "completed", "done", "ok":
		return TaskStatusSucceeded
	case "failed", "failure", "error":
		return TaskStatusFailed
	case "cancelled", "canceled", "aborted", "stopped":
		return TaskStatusCancelled
	default:
		return TaskStatusRunning // Default to running for unknown statuses
	}
}

// toLowerASCII converts ASCII letters to lowercase without allocating.
func toLowerASCII(s string) string {
	hasUpper := false
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			hasUpper = true
			break
		}
	}
	if !hasUpper {
		return s
	}
	b := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		b[i] = c
	}
	return string(b)
}
