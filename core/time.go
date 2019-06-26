package core

import (
	"context"
	"time"
)

// TimeConfiguration is used to configure time services
type TimeConfiguration struct {
	// FramesPerSecond caps frames per second that is put out
	// To unlimit, set to 0
	FramesPerSecond int
}

// NewTime creates a new time service
func NewTime(cfg TimeConfiguration) Time {
	return Time{
		fps:       cfg.FramesPerSecond,
		fpsTicker: time.NewTicker(time.Second / (time.Duration)(cfg.FramesPerSecond)),
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
