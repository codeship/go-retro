package retro

import (
	"math"
	"regexp"
	"time"
)

// RetryableError is an interface for any kind of error allowing retries
type RetryableError interface {
	Error() string
	MaxAttempts() int
	Wait(int)
}

// ErrorCreator is a type of function that creates retryable errors from standard errors
type ErrorCreator func(error) RetryableError

type backoffRetryableError struct {
	error
	maxAttempts int
}

func (err *backoffRetryableError) MaxAttempts() int {
	return err.maxAttempts
}

func (err *backoffRetryableError) Wait(count int) {
	backoffInt := int(math.Pow(float64(count), 4.0)) + count + 10
	time.Sleep(time.Duration(backoffInt) * time.Second)
}

// NewBackoffRetryableError creates a RetryableError which will retry up to maxAttempts times
// using an exponential backoff
func NewBackoffRetryableError(err error, maxAttempts int) RetryableError {
	return &backoffRetryableError{
		err,
		maxAttempts,
	}
}

type staticRetryableError struct {
	error
	maxAttempts int
	waitSeconds time.Duration
}

func (err *staticRetryableError) MaxAttempts() int {
	return err.maxAttempts
}

func (err *staticRetryableError) Wait(count int) {
	time.Sleep(err.waitSeconds * time.Second)
}

// NewStaticRetryableError creates a RetryableError which will retry up
// to maxAttempts times sleeping waitSeconds in between tries
func NewStaticRetryableError(err error, maxAttempts, waitSeconds int) RetryableError {
	return &staticRetryableError{err, maxAttempts, time.Duration(waitSeconds)}
}

// DoWithRetry will execute a function as many times as is dictated by any
// retryable errors propagated by the function
func DoWithRetry(f func() error) error {
	var err error
	try := true
	handler := &retryHandler{}
	for try {
		try, err = handler.Try(f)
	}
	return err
}

type retryHandler struct {
	attempts int
}

func (handler *retryHandler) Try(f func() error) (bool, error) {
	err := f()
	if errRetry, ok := err.(RetryableError); ok {
		retrying := handler.attempts < errRetry.MaxAttempts()
		if retrying {
			errRetry.Wait(handler.attempts)
		}
		handler.attempts++
		return retrying, errRetry
	}
	return false, err
}

// WrapRetryableError takes an error, and given a set of retryable error regexes, returns
// a suitable error type, either a standard error or a retryable error
func WrapRetryableError(err error, errorList []*regexp.Regexp, errorCreator ErrorCreator) error {
	if err == nil {
		return nil
	}
	for _, errorRegex := range errorList {
		if errorRegex.MatchString(err.Error()) {
			return errorCreator(err)
		}
	}
	return err
}
