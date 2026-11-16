package services

import (
	"context"
	"errors"
	"github.com/juggleim/jugglechat-server/storages/dbs"
	"time"
)

type ConverNotifier interface {
	NotifyConverConfig(context.Context, dbs.ConverOutbox) error
}
type ConverQueue interface {
	Claim(context.Context, int64, int64) (*dbs.ConverOutbox, error)
	Finish(context.Context, dbs.ConverOutbox, bool, int64) error
}
type ConverWorker struct {
	Store    ConverQueue
	Notifier ConverNotifier
	Timeout  time.Duration
}

var errConverNotifierUnavailable = errors.New("conversation notifier unavailable")

func (w ConverWorker) Step(ctx context.Context) error {
	if w.Store == nil || w.Notifier == nil {
		return errConverNotifierUnavailable
	}
	timeout := w.Timeout
	if timeout <= 0 || timeout > 5*time.Second {
		timeout = 5 * time.Second
	}
	job, e := w.Store.Claim(ctx, time.Now().UnixMilli(), (timeout + 30*time.Second).Milliseconds())
	if e != nil || job == nil {
		return e
	}
	sendCtx, cancel := context.WithTimeout(ctx, timeout)
	e = w.Notifier.NotifyConverConfig(sendCtx, *job)
	cancel()
	// Shutdown interruption leaves the lease recoverable, rather than claiming
	// successful delivery or blocking shutdown on an unavailable database.
	if ctx.Err() != nil {
		return ctx.Err()
	}
	return w.Store.Finish(ctx, *job, e == nil, time.Now().UnixMilli())
}

func (w ConverWorker) Run(ctx context.Context, onError func()) {
	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()
	for {
		if ctx.Err() != nil {
			return
		}
		if e := w.Step(ctx); e != nil && ctx.Err() == nil && onError != nil {
			onError()
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}
