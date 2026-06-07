package locking

import "time"

type retryTimer struct {
	timer *time.Timer
	wait  time.Duration
}

func newRetryTimer(wait time.Duration) *retryTimer {
	return &retryTimer{wait: wait}
}

func (r *retryTimer) arm() <-chan time.Time {
	if r.wait <= 0 {
		return nil
	}
	if r.timer == nil {
		r.timer = time.NewTimer(r.wait)
		return r.timer.C
	}
	r.timer.Reset(r.wait)
	return r.timer.C
}

func (r *retryTimer) stop() {
	if r.timer != nil {
		r.timer.Stop()
	}
}
