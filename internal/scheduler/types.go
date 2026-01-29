package scheduler

import (
	"context"
	"time"
)

type Job struct {
	ID       int64                                      `json:"id"`
	Name     string                                     `json:"name"`
	Priority int                                        `json:"priority"`
	Duration time.Duration                              `json:"duration"`
	Execute  func(ctx context.Context, d time.Duration) `json:"-"`
}
