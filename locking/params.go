package locking

import "time"

type LockParams struct {
	TimeToLive time.Duration
	WaitTime   time.Duration
	MaxRetries int
}

var DefaultLockParams = NewLockParams(2*time.Second, 100*time.Millisecond, 10)

func NewLockParams(timeToLive time.Duration, waitTime time.Duration, maxRetries int) LockParams {
	return LockParams{
		TimeToLive: timeToLive,
		WaitTime:   waitTime,
		MaxRetries: maxRetries,
	}
}
