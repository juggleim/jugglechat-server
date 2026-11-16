package services

import (
	"context"
	"errors"
	"github.com/juggleim/jugglechat-server/storages/dbs"
	"testing"
	"time"
)

type workerFixture struct {
	job               *dbs.ConverOutbox
	finished, success bool
}

func (f *workerFixture) Claim(context.Context, int64, int64) (*dbs.ConverOutbox, error) {
	return f.job, nil
}
func (f *workerFixture) Finish(_ context.Context, _ dbs.ConverOutbox, success bool, _ int64) error {
	f.finished = true
	f.success = success
	return nil
}

type notifierFunc func(context.Context, dbs.ConverOutbox) error

func (f notifierFunc) NotifyConverConfig(c context.Context, j dbs.ConverOutbox) error { return f(c, j) }
func TestConverWorkerTimeoutAndRetry(t *testing.T) {
	f := &workerFixture{job: &dbs.ConverOutbox{EventID: "fixture"}}
	w := ConverWorker{Store: f, Timeout: 10 * time.Millisecond, Notifier: notifierFunc(func(c context.Context, j dbs.ConverOutbox) error { <-c.Done(); return c.Err() })}
	if e := w.Step(context.Background()); e != nil {
		t.Fatal(e)
	}
	if !f.finished || f.success {
		t.Fatal("timeout marked sent")
	}
	w.Notifier = notifierFunc(func(context.Context, dbs.ConverOutbox) error { return nil })
	if e := w.Step(context.Background()); e != nil || !f.success {
		t.Fatal(e)
	}
	w.Notifier = nil
	if !errors.Is(w.Step(context.Background()), errConverNotifierUnavailable) {
		t.Fatal("missing notifier succeeds")
	}
}

func TestConverWorkerShutdownLeavesRecoverableLease(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	f := &workerFixture{job: &dbs.ConverOutbox{EventID: "fixture"}}
	w := ConverWorker{Store: f, Notifier: notifierFunc(func(c context.Context, _ dbs.ConverOutbox) error { cancel(); <-c.Done(); return c.Err() })}
	if !errors.Is(w.Step(ctx), context.Canceled) || f.finished {
		t.Fatal("shutdown finalized interrupted delivery")
	}
	done := make(chan struct{})
	go func() { w.Run(ctx, nil); close(done) }()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("worker failed to stop")
	}
}
