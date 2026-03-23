package backoff

import (
	"time"
)

type SimpleExponentialBackoff struct {
	prev uint64
	base uint64
}

func NewExponentialBackoff(option *Option) Backoff {
	backoff := &SimpleExponentialBackoff{
		prev: 1,
		base: 200,
	}
	if option != nil {
		backoff.base = option.BaseInMilliseconds
	}
	return backoff
}

func (backoff *SimpleExponentialBackoff) Reset() {
	backoff.prev = 1
}

func (backoff *SimpleExponentialBackoff) Next() time.Duration {
	next := backoff.base * backoff.prev
	backoff.prev = 2 * backoff.prev
	return time.Duration(next) * time.Millisecond
}
