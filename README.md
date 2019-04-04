## Retro - Retryable errors in Golang

:warning: **Note:** This library is no longer maintained :warning:

[ ![Codeship Status for codeship/go-retro](https://codeship.com/projects/7a5b0350-1b79-0134-e8c7-265a91e3d879/status?branch=master)](https://codeship.com/projects/159666)

Retry allows you to wrap failure prone sections of code in retry blocks, and conditionally retry different types of errors. To retry an error within a block, simply return a custom error type matching the provided interface. By default, wrapping code in a retro block will have no effect.

### Usage

```
// Wrap your code block with retro
finalErr := retro.DoWithRetry(func() error{
    // Hit an external API
    return DoMyFunc(args)
})
```

Any errors matching the `RetryableError` interface bubbled up will automatically retry the code block.


For more detailed examples see the [examples folder](https://github.com/codeship/go-retro/blob/master/examples).

Independent sections of code within a chain can be wrapped conditionally in retro blocks as needed. You should always try and keep retry blocks as small as possible to reduce code and requests being re-run.

### Retrying

You can use two types of provided error retry patterns, or create your own:

#### Static retryable errors

These errors retry X times every Y seconds. To create a static retryable error use `retro.NewStaticRetryableError(err error, maxAttempts, waitSeconds int)`. For generic errors you can reference a shared variable, for dynamic errors you can create an error with the relevant base error message each time. Error retrying state is stored in the retro loop, not the RetryableError.

#### Backoff retryable errors

These errors use quadrilateral equation based on Sidekiq retries to increasingly space out subsequent attempts. To create a backoff retryable error use `retro.NewBackoffRetryableError(err error, maxAttempts int)`. For generic errors you can reference a shared variable, for dynamic errors you can create an error with the relevant base error message each time. Error retrying state is stored in the retro loop, not the RetryableError.

### Retry attempts

The retro retry loop will keep track of how many times it has looped. Any time it gets an error it compares the state against the allowed retry conditions for the latest error. This means that should the loop initially retry with an error allowing ten retries, if the second error indicates only two retries or fewer are allowed, then the loop will no longer retry since it has already retried twice.
