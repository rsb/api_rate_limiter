# Architecture
## Rate Limiting Algorithm 
There are many ways to code rate limiting functionality into a web server:
- Fixed Window
- Sliding Window
- Token Bucket
- Leaky Token Bucket 

I chose a variation of the `Token Bucket` found at [limits](foundation/limits/limits.go).
The `Token Bucket Algorithm` used a fixed capacity bucket where tokens are added
at fixed rate. Each token represents in our case a single request against the 
web server. We `take` a token from the bucket for every request that is serviced.
if we run out of requests then we have reached our limit an return `429 too many requests` 

In my code I have the following:

### Config
Handles configuration to allow us to control rate limiting
- `limit` is the rate limit or the number of tokens we will start with out bucket
- `Interval` is the duration the limit is measured against. like `1min` or `12hours`
- `TTLInterval` is used to determine when to kick off clean up `memory management`
- `MinTTL` is used to determine when to delete the entry
- `InitialSize` controls the first allocation of the map that holds the buckets
```go
type Config struct {
  Limit       uint64
  Interval    time.Duration
  TTLInterval time.Duration
  MinTTL      time.Duration
  InitialSize int
}
```

### Bucket
The bucket holds the metadata about the rate limit for a given key. It is also responsible
for filling itself up once the interval has expired.

```go
type Bucket struct {
	startTime       uint64
	maxTokens       uint64
	interval        time.Duration
	availableTokens uint64
	lastTick        uint64
	lock            sync.Mutex
}

```

### MemoryStore
In memory storage client to control the behavior of the limiter. It has the 
following interface

```go
type RateLimiter interface {
	Take(key string) (RateInfo, error)
	Get(key string) (limit uint64, remaining uint64)
	Set(key string, tokens uin64, interval time.Duration) error
	Close() error
	GarbageCollector()
}
```

### Take
Looks first for a quick read only using an `RLock` if the key was already added
then this path is efficient. It allows read access to still be available. If not
found then it will take a full lock, check once again if the key was added concurrently
otherwise it will add a new bucket to the data structure. Besides `GarageCollector` this
is the only other function required for middleware


### Get
Uses only a read lock to pull the `token limit and remaining tokens` from the bucket

### Set
Adds a new bucket for a given key

### GarbageCollector
Continually iterates over the map and purges on the provided ttl info. This is run
on a separate go routine and is started up when the middleware in created


# Example Application
This is quick overview of the code layout for the example app. 

## App
The app folder handles all the application concerns required to start and stop and inspect it.
This includes the following sub folders:

- `api` holds the handlers and middle ware. The handlers are the main entry point for the app.
- `cli` holds a [cobra](https://github.com/spf13/cobra) command line app used to start the server
- `conf` is package used to hold `env,cli,config file` setting for the application
- `construct` is package the consumes config data to build (`construct`) dependencies for the app


## Business
This holds the business logic for the application. Typically, it would have the following
sub packages:
- `platform` holds packages for concrete adapters like db, rest api etc..
- `features` each app feature is its own package and defines an interface for that `platform` adapters implement the construct package builds the dependencies and injects the into the `features`
- `data` package used to handle data serialization and presentation for business types

## Foundation
This package holds packages that could be used across many microservices and maturing before moving to a organizations gokit

## Infra
Holds docker, kubernetes, terraform files for creating the necessary resources for this microservice

## Tests
integration tests. Only a very basic example for this.
             