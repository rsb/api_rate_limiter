# API Rate Limiting Example
Example implementation of limiting api call for rest service in golang.

The rate limiter will monitor the number of requests per window of time which 
is determined through configuration. I f the request count exceeds the rate
limiters max number then the call will be rejected returning a http status

```
429 Too Man Requests
```

This particular implementation will limit requests by client IP address