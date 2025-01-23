package locking

import "time"

type LockParams struct {
	TimeToLive      time.Duration
	WaitTime        time.Duration
	MaxRetries      int
	InitialWaitTime time.Duration
}

var DefaultLockParams = NewLockParams(2*time.Second, 100*time.Millisecond, 10, 100*time.Millisecond)

func ReplaceDefaultWaitTime(waitTime time.Duration) LockParams {
	DefaultLockParams.WaitTime = waitTime
	return DefaultLockParams
}

func ReplaceDefaultTimeToLive(timeToLive time.Duration) LockParams {
	DefaultLockParams.TimeToLive = timeToLive
	return DefaultLockParams
}

func ReplaceDefaultMaxRetries(maxRetries int) LockParams {
	DefaultLockParams.MaxRetries = maxRetries
	return DefaultLockParams
}

func NewLockParams(timeToLive time.Duration, waitTime time.Duration, maxRetries int, initialWaitTime time.Duration) LockParams {
	return LockParams{
		TimeToLive:      timeToLive,
		WaitTime:        waitTime,
		MaxRetries:      maxRetries,
		InitialWaitTime: initialWaitTime,
	}
}
