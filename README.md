## Retro - Retryable errors in Golang

[ ![Codeship Status for codeship/retro](https://codeship.com/projects/9b134140-1aff-0134-ce79-0e8ad2af7d49/status?branch=master)](https://codeship.com/projects/159527)

Retry allows you to wrap failure prone sections of code in retry blocks, and conditionally retry different types of errors. To retry an error within a block, simply return a custom error type matching the provided interface. By default, wrapping code in a retro block will have no effect.

Example:

```
ErrNetwork := retro.NewStaticRetryableError(errors.New("Network error!", 3, 10))
userID := "abc123"

finalErr := retro.DoWithRetry(func() error{
    // Get a user record, should not fail, do not retry
    user, err := db.GetUser(userID)
    if err != nil {
        return err
    }

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
    return nil
})
```

Independent sections of code within a chain can be wrapped conditionally in retro blocks as needed. You should always try and keep retry blocks as small as possible to reduce code and requests being re-run.


