package core

import (
	"context"
	"time"
)

// NewTime creates a new time service
func NewTime(cfg TimeConfiguration) Time {
	var interval time.Duration
	if cfg.FramesPerSecond == 0 {
		interval = time.Nanosecond
	} else {
		interval = time.Second / (time.Duration)(cfg.FramesPerSecond)
	}

	return Time{
		fps:       cfg.FramesPerSecond,
		fpsTicker: time.NewTicker(interval),
	}
}

// Time containes all the time services and tickers
type Time struct {
	ctx context.Context

	fps       int
	fpsTicker *time.Ticker
}

// Fps gets the set frames per second
func (t *Time) Fps() int {
	return t.fps
}

// FpsTicker gets the initialized fps ticker
func (t *Time) FpsTicker() *time.Ticker {
	return t.fpsTicker
}
