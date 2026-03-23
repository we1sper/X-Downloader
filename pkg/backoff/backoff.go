package backoff

import (
	"fmt"
	"time"
)

var factories = map[string]func(option *Option) Backoff{}

func init() {
	Register("SimpleExponential", NewExponentialBackoff)
}

type Option struct {
	BaseInMilliseconds uint64
}

type Backoff interface {
	Reset()
	Next() time.Duration
}

func Register(name string, factory func(option *Option) Backoff) {
	factories[name] = factory
}

func Get(name string, option *Option) (Backoff, error) {
	backoff, ok := factories[name]
	if !ok {
		return nil, fmt.Errorf("factory of backoff '%s' not found", name)
	}
	return backoff(option), nil
}
