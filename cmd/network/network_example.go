package main

import (
	"errors"
	"fmt"
	"math/rand"
	"os"

	"github.com/codeship/go-retro"
)

var (
	// ErrInvalidVersion is returned when the wrong version is provided. This error is not retryable
	ErrInvalidVersion = errors.New("error: invalid version")

	// ErrNetwork is returned when a network error occurs. This error is retryable and will retry 5 times sleeping for 3 seconds each time
	ErrNetwork = retro.NewStaticRetryableError(errors.New("error: failed to connect"), 5, 3)

	// ErrNotReady is returned when a resource is not ready. This error is retryable and will retry 5 times with an incremental backoff
	ErrNotReady = retro.NewBackoffRetryableError(errors.New("error: resource not ready"), 5)
)

// Gets a server version by id, and tries to use it until it is ready
func main() {
	serverID := "abc123"
	var version string
	err := retro.DoWithRetry(func() error {
		var e error
		version, e = getServer(serverID)
		if e != nil {
			return e
		}
		fmt.Printf("Server version %s\n", version)
		return nil
	})
	if err != nil {
		fmt.Printf("FATAL: Failed to get server info %s\n", err.Error())
		os.Exit(1)
	}

	err = retro.DoWithRetry(func() error {
		return useServer(serverID, version)
	})
	if err != nil {
		fmt.Printf("FATAL: Failed to use server %s\n", err.Error())
		os.Exit(1)
	}

}

func getServer(id string) (string, error) {
	// return makeCurlRequestTODO()
	if maybeFail() {
		// return a intermittent network error which we can retry
		return "", ErrNetwork
	} else if maybeFail() {
		// return an incompatibility error which we cannot
		return "", ErrInvalidVersion
	}
	return "1", nil
}
func useServer(id, version string) error {
	// return makeCurlRequestTODO()
	if maybeFail() {
		// return an error which should resolve after an unknown amount of time
		return ErrNotReady
	}
	return nil
}

func maybeFail() bool {
	return rand.Intn(5) == 0
}
