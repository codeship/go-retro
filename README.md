## Retro - Retryable errors in Golang

[ ![Codeship Status for codeship/go-retro](https://codeship.com/projects/7a5b0350-1b79-0134-e8c7-265a91e3d879/status?branch=master)](https://codeship.com/projects/159666)

Retry allows you to wrap failure prone sections of code in retry blocks, and conditionally retry different types of errors. To retry an error within a block, simply return a custom error type matching the provided interface. By default, wrapping code in a retro block will have no effect.

### Example

```
ErrNetwork := retro.NewStaticRetryableError(errors.New("Network error!"), 3, 10)

finalErr := retro.DoWithRetry(func() error{
    // Hit an external API
    sales, err := api.GetSalesForUser(user)
    if err != nil {
        // if its an EOF error
        if err.Error() == "EOF" {
            // return a generic retryable error
            return ErrNetwork
        } else {
          // otherwise return a custom retryable error retrying 5 times, once every 30 seconds
          return retro.NewStaticRetryableError(err, 5, 30)
        }
    }

    // Store sales in DB, do not retry
    if err := db.SaveSales(sales); err != nil {
        return err
    }

    invoices, err := api.GetInvoicesForSales(sales)
    if err != nil {
        // if its an EOF error
        if err.Error() == "EOF" {
            // return a generic retryable error
            return ErrNetwork
        } else {
          // otherwise return a custom retryable error retrying 5 times using the Sidekiq backoff equation
          return retro.NewBackoffRetryableError(err, 5)
        }
    }
    return nil
})
```

Independent sections of code within a chain can be wrapped conditionally in retro blocks as needed. You should always try and keep retry blocks as small as possible to reduce code and requests being re-run.

### Retrying

You can use two types of provided error retry patterns, or create your own:

#### Static retryable errors

These errors retry X times every Y seconds. To create a static retryable error use `retro.NewStaticRetryableError(err error, maxAttempts, waitSeconds int)`. For generic errors you can reference a shared variable, for dynamic errors you can create an error with the relevant base error message each time. Error retrying state is stored in the retro loop, not the RetryableError.

#### Backoff retryable errors

These errors use quadrilateral equation based on Sidekiq retries to increasingly space out subsequent attempts. To create a backoff retryable error use `retro.NewBackoffRetryableError(err error, maxAttempts int)`. For generic errors you can reference a shared variable, for dynamic errors you can create an error with the relevant base error message each time. Error retrying state is stored in the retro loop, not the RetryableError.
