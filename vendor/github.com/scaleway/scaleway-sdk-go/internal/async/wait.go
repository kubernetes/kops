package async

import (
	"fmt"
	"time"
)

var (
	defaultInterval = time.Second
	defaultTimeout  = time.Minute * 5
)

type IntervalStrategy func() <-chan time.Time

// WaitSyncConfig defines the waiting options.
type WaitSyncConfig struct {
	Get              func() (value any, isTerminal bool, err error)
	IntervalStrategy IntervalStrategy
	Timeout          time.Duration
}

// LinearIntervalStrategy defines a linear interval duration.
func LinearIntervalStrategy(interval time.Duration) IntervalStrategy {
	return func() <-chan time.Time {
		return time.After(interval)
	}
}

// FibonacciIntervalStrategy defines an interval duration who follow the Fibonacci sequence.
func FibonacciIntervalStrategy(base time.Duration, factor float32) IntervalStrategy {
	var x, y float32 = 0, 1

	return func() <-chan time.Time {
		x, y = y, x+(y*factor)
		return time.After(time.Duration(x) * base)
	}
}

// WaitSync waits and returns when a given stop condition is true or if an error occurs.
func WaitSync(config *WaitSyncConfig) (terminalValue any, err error) {
	// initialize configuration
	if config.IntervalStrategy == nil {
		config.IntervalStrategy = LinearIntervalStrategy(defaultInterval)
	}

	if config.Timeout == 0 {
		config.Timeout = defaultTimeout
	}

	timeoutCh := time.After(config.Timeout)

	for {
		// get the payload
		value, stopCondition, err := config.Get()
		if err != nil {
			return nil, err
		}
		if stopCondition {
			return value, nil
		}

		select {
		case <-timeoutCh:
			return nil, fmt.Errorf("timeout after %v", config.Timeout)
		case <-config.IntervalStrategy(): // Sleep before next get() call.
		}
	}
}
