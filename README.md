# API Rate Limiting Example
Example implementation of limiting api call for rest service in golang.

The rate limiter will monitor the number of requests per window of time which 
is determined through configuration. I f the request count exceeds the rate
limiters max number then the call will be rejected returning a http status

```
429 Too Man Requests
```

This particular implementation will limit requests by client IP address

## Usage
To run the example you have four options:

### Using Go run
You can use the cli by using `go run` on the terminal. A developer might do
a quick check using this method, certainly not a production workflow.
```shell
go run app/cli/limits/main.go api serve
```

### Manually build go
You can manually compile the cli app and run the web server.

```shell
 make build
 ./limits api serve
```

### Run docker
```shell
make docker-limit-api
make docker-run-limit-api
```

You can manually stop or kill it using `docker stop/kill [containerID]`. the last
make command will output the container id

### Run Kubernetes using kind
Currently, only kind is support as this is just an example.
```shell
make kind-up
make docker-limits-api
make kind-load
make kind-apply

```
After that the endpoint should be reachable at 
```shell
http://locahost:3000/ping
```