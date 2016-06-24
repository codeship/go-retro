package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/codeship/go-retro"
	"github.com/satori/go.uuid"
)

var (
	// ErrNilMap is returned when an uninitialized data map is provided
	ErrNilMap = errors.New("error: data map was nil")

	// ErrKeyExists is returned when a key is re-used. This error is retryable and will retry 5 times without sleeping
	ErrKeyExists = retro.NewStaticRetryableError(errors.New("error: key exists"), 5, 0)
)

// Stores all ints from 0-99 under random unique keys in a map
func main() {
	data := map[string]int{}

	for i := 0; i < 100; i++ {
		err := retro.DoWithRetry(func() error {
			return storeInt(data, i)
		})
		if err != nil {
			fmt.Printf("FATAL: Failed to store %d: %s\n", i, err.Error())
			os.Exit(1)
		}
	}

}

func storeInt(data map[string]int, i int) error {
	if data == nil {
		return ErrNilMap
	}

	key := uuid.NewV4().String()
	_, ok := data[key]
	if ok {
		return ErrKeyExists
	}
	data[key] = i
	return nil
}
