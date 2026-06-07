package locking

import "time"

type lockEntry struct {
	ch      chan struct{}
	token   uint64
	waiters int
	timer   *time.Timer
}

func (l *lockEntry) drainCh() {
	if l == nil {
		return
	}
	select {
	case <-l.ch:
	default:
	}
}

func (l *lockEntry) releaseIfToken(token uint64) {
	if l.token != token {
		return
	}
	l.releaseToken()
}

func (l *lockEntry) releaseToken() {
	l.drainCh()
	l.token++
	if l.timer != nil {
		l.timer.Stop()
		l.timer = nil
	}
}
