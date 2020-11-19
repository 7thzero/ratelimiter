# (go) ratelimiter
Sample rate limiter for go HTTP services. This a simple implementation that has not been tested in a production environment and is meant to illustrate the concept.

See examples/sample-ratelimiter.go for a demonstration

# How to use this:
- Build an HTTP router that incorporates the rate limiter:

```
type SampleRouter struct{
   	throttler map[string]ratelimiter.RateLimit
   }
```

- Then, implement a function to use the rate limiter:

```
	// Check if we should limit/block the connection attempt
	host := strings.Split(request.RemoteAddr, ":")[0]
	if sr.dosShield(host){
		writer.WriteHeader(http.StatusTooManyRequests)
		writer.Write([]byte(`{"message": "Rate limit exceeded"}`))
		return
	}
```
```
//
// leverages the rate limiter to check if DOS is suspected
func (sr *SampleRouter) dosShield(host string) bool{
	// Ensure we have an entry for the host if it doesn't already exist
	if _, exists := sr.throttler[host]; !exists{
		limiter := ratelimiter.RateLimit{}
		limiter.Init(15, 65)
		sr.throttler[host] = limiter
	}

	hostLimit := sr.throttler[host]				// Get the throttler for this host
	isRateLimited := hostLimit.IsRateLimited()	// Records the access attempt and checks if this connection should be throttled
	sr.throttler[host] = hostLimit				// Re-assign the RateLimiter to record the updated access attempt
	return isRateLimited
}
```

