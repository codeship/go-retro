package retro

import (
	"errors"
	"regexp"
	"testing"

	"github.com/codeship/janus/src/mt"
)

type stubRetryableError struct {
	error
	maxAttempts int
	sleepFunc   func(int)
}

func (err *stubRetryableError) MaxAttempts() int {
	return err.maxAttempts
}

func (err *stubRetryableError) Wait(count int) {
	err.sleepFunc(count)
}

func TestRetryHandlerTryWithSuccess(t *testing.T) {
	handler := &retryHandler{}
	successfulFunc := func() error {
		return nil
	}
	retry, err := handler.Try(successfulFunc)
	mt.Refute(t, retry)
	mt.AssertEqual(t, err, nil)
}

func TestRetryHandlerTryNonRetryable(t *testing.T) {
	handler := &retryHandler{}
	nonRetryableError := errors.New("testErr")
	nonRetryableFunc := func() error {
		return nonRetryableError
	}
	retry, err := handler.Try(nonRetryableFunc)
	mt.Refute(t, retry)
	mt.AssertEqual(t, err, nonRetryableError)
}

func TestRetryHandlerTryRetryable(t *testing.T) {
	handler := &retryHandler{}
	receivedCounts := []int{}
	baseError := errors.New("foobar")
	retryableError := &stubRetryableError{
		baseError,
		2,
		func(count int) {
			receivedCounts = append(receivedCounts, count)
		},
	}
	retryableFunc := func() error {
		return retryableError
	}

	retry, err := handler.Try(retryableFunc)
	mt.Assert(t, retry)
	mt.AssertEqual(t, receivedCounts[0], 0)
	mt.AssertEqual(t, len(receivedCounts), 1)
	mt.AssertEqual(t, err, retryableError)

	retry, err = handler.Try(retryableFunc)
	mt.Assert(t, retry)
	mt.AssertEqual(t, receivedCounts[1], 1)
	mt.AssertEqual(t, len(receivedCounts), 2)
	mt.AssertEqual(t, err, retryableError)

	retry, err = handler.Try(retryableFunc)
	mt.Refute(t, retry)
	mt.AssertEqual(t, len(receivedCounts), 2)
	mt.AssertEqual(t, err, retryableError)
}

func TestRetryHandlerTryRetryableEventualSuccess(t *testing.T) {
	handler := &retryHandler{}
	receivedCounts := []int{}
	baseError := errors.New("foobar")
	retryableError := &stubRetryableError{
		baseError,
		2,
		func(count int) {
			receivedCounts = append(receivedCounts, count)
		},
	}
	returnIndex := 0
	returns := []error{retryableError, nil}
	retryableFunc := func() error {
		returnIndex++
		return returns[returnIndex-1]
	}

	retry, err := handler.Try(retryableFunc)
	mt.Assert(t, retry)
	mt.AssertEqual(t, receivedCounts[0], 0)
	mt.AssertEqual(t, len(receivedCounts), 1)
	mt.AssertEqual(t, err, retryableError)

	retry, err = handler.Try(retryableFunc)
	mt.Refute(t, retry)
	mt.AssertEqual(t, len(receivedCounts), 1)
	mt.AssertEqual(t, err, nil)
}

func TestRetryHandlerTryRetryableEventualFailure(t *testing.T) {
	handler := &retryHandler{}
	receivedCounts := []int{}
	baseError := errors.New("foobar")
	retryableError := &stubRetryableError{
		baseError,
		2,
		func(count int) {
			receivedCounts = append(receivedCounts, count)
		},
	}
	returnIndex := 0
	returns := []error{retryableError, baseError}
	retryableFunc := func() error {
		returnIndex++
		return returns[returnIndex-1]
	}

	retry, err := handler.Try(retryableFunc)
	mt.Assert(t, retry)
	mt.AssertEqual(t, receivedCounts[0], 0)
	mt.AssertEqual(t, len(receivedCounts), 1)
	mt.AssertEqual(t, err, retryableError)

	retry, err = handler.Try(retryableFunc)
	mt.Refute(t, retry)
	mt.AssertEqual(t, len(receivedCounts), 1)
	mt.AssertEqual(t, err, baseError)
}

func TestDoWithRetry(t *testing.T) {
	timesCalled := 0
	retryableFunc := func() error {
		timesCalled++
		return nil
	}

	err := DoWithRetry(retryableFunc)
	mt.AssertEqual(t, err, nil)
	mt.AssertEqual(t, timesCalled, 1)
}

func TestDoWithRetryEventualFailure(t *testing.T) {
	baseError := errors.New("foobar")
	retryableError := &stubRetryableError{
		baseError,
		2,
		func(count int) {},
	}
	returnIndex := 0
	returns := []error{retryableError, baseError}
	retryableFunc := func() error {
		returnIndex++
		return returns[returnIndex-1]
	}

	err := DoWithRetry(retryableFunc)
	mt.AssertEqual(t, err, baseError)
}

func TestDoWithRetryEventualSuccess(t *testing.T) {
	baseError := errors.New("foobar")
	retryableError := &stubRetryableError{
		baseError,
		2,
		func(count int) {},
	}
	returnIndex := 0
	returns := []error{retryableError, nil}
	retryableFunc := func() error {
		returnIndex++
		return returns[returnIndex-1]
	}

	err := DoWithRetry(retryableFunc)
	mt.AssertEqual(t, err, nil)
}

func TestWrapRetryableErrors(t *testing.T) {
	err := errors.New("foobar")
	errorList := []*regexp.Regexp{
		regexp.MustCompile("foo"),
	}
	err1 := WrapRetryableError(err, errorList, func(e error) RetryableError {
		return NewBackoffRetryableError(e, 1)
	})

	_, ok := err1.(RetryableError)
	mt.AssertEqual(t, err.Error(), err1.Error())
	mt.Assert(t, ok)
}

func TestWrapRetryableFailure(t *testing.T) {
	err := errors.New("foobar")
	errorList := []*regexp.Regexp{
		regexp.MustCompile("fooop"),
	}
	err1 := WrapRetryableError(err, errorList, func(e error) RetryableError {
		return NewBackoffRetryableError(e, 1)
	})

	_, ok := err1.(RetryableError)
	mt.AssertEqual(t, err.Error(), err1.Error())
	mt.Refute(t, ok)
}

func TestWrapRetryableNil(t *testing.T) {
	errorList := []*regexp.Regexp{
		regexp.MustCompile("fooop"),
	}
	err1 := WrapRetryableError(nil, errorList, func(e error) RetryableError {
		return NewBackoffRetryableError(e, 1)
	})

	mt.AssertEqual(t, err1, nil)
}
