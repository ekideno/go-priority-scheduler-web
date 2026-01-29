package scheduler

import (
	"context"
	"time"
)

type Job struct {
	ID       int64
	Name     string
	Priority int
	Duration time.Duration
	Execute  func(ctx context.Context, d time.Duration) `json:"-"`
}
