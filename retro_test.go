package retro

import (
	"errors"
	"regexp"
	"testing"
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
	if retry {
		t.Errorf("retry was true, expected false")
	}
	if err != nil {
		t.Errorf("error was %s, expected nil", err.Error())
	}
}

func TestRetryHandlerTryNonRetryable(t *testing.T) {
	handler := &retryHandler{}
	nonRetryableError := errors.New("testErr")
	nonRetryableFunc := func() error {
		return nonRetryableError
	}
	retry, err := handler.Try(nonRetryableFunc)
	if retry {
		t.Errorf("retry was true, expected false")
	}
	if err != nonRetryableError {
		t.Errorf("error was %s, expected %s", err.Error(), nonRetryableError.Error())
	}
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
	if !retry {
		t.Errorf("retry was false, expected true")
	}
	if receivedCounts[0] != 0 {
		t.Errorf("receivedCounts[0] was %d, expected %d", receivedCounts[0], 0)
	}
	if len(receivedCounts) != 1 {
		t.Errorf("len(receivedCounts) was %d, expected %d", len(receivedCounts), 1)
	}
	if err != retryableError {
		t.Errorf("error was %s, expected %s", err.Error(), retryableError.Error())
	}
	retry, err = handler.Try(retryableFunc)
	if !retry {
		t.Errorf("retry was false, expected true")
	}
	if receivedCounts[1] != 1 {
		t.Errorf("receivedCounts[1] was %d, expected %d", receivedCounts[1], 1)
	}
	if len(receivedCounts) != 2 {
		t.Errorf("len(receivedCounts) was %d, expected %d", len(receivedCounts), 2)
	}
	if err != retryableError {
		t.Errorf("error was %s, expected %s", err.Error(), retryableError.Error())
	}
	retry, err = handler.Try(retryableFunc)
	if retry {
		t.Errorf("retry was true, expected false")
	}

	if len(receivedCounts) != 2 {
		t.Errorf("len(receivedCounts) was %d, expected %d", len(receivedCounts), 2)
	}
	if err != retryableError {
		t.Errorf("error was %s, expected %s", err.Error(), retryableError.Error())
	}
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
	if !retry {
		t.Errorf("retry was false, expected true")
	}
	if receivedCounts[0] != 0 {
		t.Errorf("receivedCounts[0] was %d, expected %d", receivedCounts[0], 0)
	}

	if len(receivedCounts) != 1 {
		t.Errorf("len(receivedCounts) was %d, expected %d", len(receivedCounts), 1)
	}
	if err != retryableError {
		t.Errorf("error was %s, expected %s", err.Error(), retryableError.Error())
	}
	retry, err = handler.Try(retryableFunc)
	if retry {
		t.Errorf("retry was true, expected false")
	}
	if len(receivedCounts) != 1 {
		t.Errorf("len(receivedCounts) was %d, expected %d", len(receivedCounts), 1)
	}
	if err != nil {
		t.Errorf("error was %s, expected nil", err.Error())
	}
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
	if !retry {
		t.Errorf("retry was false, expected true")
	}
	if receivedCounts[0] != 0 {
		t.Errorf("receivedCounts[0] was %d, expected %d", receivedCounts[0], 0)
	}
	if len(receivedCounts) != 1 {
		t.Errorf("len(receivedCounts) was %d, expected %d", len(receivedCounts), 1)
	}
	if err != retryableError {
		t.Errorf("error was %s, expected %s", err.Error(), retryableError.Error())
	}
	retry, err = handler.Try(retryableFunc)
	if retry {
		t.Errorf("retry was true, expected false")
	}
	if len(receivedCounts) != 1 {
		t.Errorf("len(receivedCounts) was %d, expected %d", len(receivedCounts), 1)
	}
	if err != baseError {
		t.Errorf("error was %s, expected %s", err.Error(), baseError.Error())
	}
}

func TestDoWithRetry(t *testing.T) {
	timesCalled := 0
	retryableFunc := func() error {
		timesCalled++
		return nil
	}

	err := DoWithRetry(retryableFunc)
	if err != nil {
		t.Errorf("error was %s, expected nil", err.Error())
	}
	if timesCalled != 1 {
		t.Errorf("timesCalled was %d, expected %d", timesCalled, 1)
	}
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
	if err != baseError {
		t.Errorf("error was %s, expected %s", err.Error(), baseError.Error())
	}
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
	if err != nil {
		t.Errorf("error was %s, expected nil", err.Error())
	}
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
	if err.Error() != err1.Error() {
		t.Errorf("error was %s, expected %s", err.Error(), err1.Error())
	}
	if !ok {
		t.Errorf("ok was false, expected true")
	}
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
	if err != err1 {
		t.Errorf("error was %s, expected %s", err.Error(), err1.Error())
	}
	if ok {
		t.Errorf("ok was true, expected false")
	}
}

func TestWrapRetryableNil(t *testing.T) {
	errorList := []*regexp.Regexp{
		regexp.MustCompile("fooop"),
	}
	err1 := WrapRetryableError(nil, errorList, func(e error) RetryableError {
		return NewBackoffRetryableError(e, 1)
	})

	if err1 != nil {
		t.Errorf("error was %s, expected nil", err1.Error())
	}
}
