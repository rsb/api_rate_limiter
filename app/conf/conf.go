// Package conf is responsible defining application configuration for
// Features, Adapters/Ports, Middleware or any other system that
// requires configuration.
package conf

import (
	"github.com/gofiber/fiber/v2"
	"github.com/mitchellh/go-homedir"
	"github.com/rsb/failure"
	"time"
)

type LimiterAPI struct {
	Version
	Kubernetes
	API
}

type Version struct {
	Build string `conf:"env:API_BUILD_VERSION, cli:api-build-version, cli-u:version of the web api"`
	Desc  string `conf:"env:API_BUILD_DESC,    cli:api-build-desc,    cli-u:summary of the build"`
}

type API struct {
	Host                   string        `conf:"env:API_HOST, cli:api-host, default:0.0.0.0:3000, cli-u:web api host"`
	DebugHost              string        `conf:"env:API_DEBUG_HOST, cli:debug-host, default:0.0.0.0:4000, cli-u:debug host"`
	IsCaseSensitive        bool          `conf:"env:API_ROUTE_CASE_SENSITIVE, cli:api-route-case-sensitive, default:false, cli-u:will routes be case sensitive"`
	IsETag                 bool          `conf:"env:API_ETAG, cli:api-etag, default:false, cli-u:enable/disable etag header generation"`
	ReadTimeout            time.Duration `conf:"env:API_READ_TIMEOUT,cli:api-read-timeout, default:5s"`
	WriteTimeout           time.Duration `conf:"env:API_WRITE_TIMEOUT,cli:api-write-timeout, default:20s"`
	IdleTimeout            time.Duration `conf:"env:API_IDLE_TIMEOUT, cli:api-idle-timeout, default:120s"`
	ShutdownTimeout        time.Duration `conf:"env:API_SHUTDOWN_TIMEOUT,cli:api-shutdown-timeout, default:20s"`
	RateLimit              uint64        `conf:"env:API_RATE_LIMIT,cli:api-rate-limit, default:10"`
	RateLimitInterval      time.Duration `conf:"env:API_RATE_LIMIT_INTERVAL,cli:api-rate-limit-interval, default:60s"`
	RateLimitCleanStale    time.Duration `conf:"env:API_RATE_CLEAN_STALE, cli:api-rate-limit-clean-stale, default:6h"`
	RateLimitCleanInactive time.Duration `conf:"env:API_RATE_CLEAN_INACTIVE, cli:api-rate-limit-clean-inactive, default:12h"`
}

func (a API) NewFiberConfig() fiber.Config {
	config := fiber.Config{
		IdleTimeout:   a.IdleTimeout,
		ReadTimeout:   a.ReadTimeout,
		WriteTimeout:  a.WriteTimeout,
		CaseSensitive: a.IsCaseSensitive,
		ETag:          a.IsETag,
	}

	return config
}

type HTTPClient struct {
	Timeout            time.Duration `conf:"default: 5s,  env:LOLA_HTTP_CLIENT_TIMEOUT, cli:http-client-timeout, cli-u:timeout for http clients"`
	MaxIdleConn        int           `conf:"default: 100, env:LOLA_HTTP_CLIENT_MAX_IDLE_CONN, cli:http-client-max-idle-con, cli-u:http client max idle connections"`
	MaxConnPerHost     int           `conf:"default: 100, env:LOLA_HTTP_CLIENT_MAX_CONN_PER_HOST, cli:http-client-max-con-per-host, cli-u:http client max connection per host"`
	MaxIdleConnPerHost int           `conf:"default: 100, env:LOLA_HTTP_CLIENT_MAX_IDLE_PER_HOST, cli:http-client-max-idle-per-host, cli-u:http client max idle connections per host"`
}

type Kubernetes struct {
	Pod       string `conf:"env:KUBERNETES_PODNAME"`
	PodIP     string `conf:"env:KUBERNETES_NAMESPACE_POD_IP"`
	Node      string `conf:"env:KUBERNETES_NODENAME"`
	Namespace string `conf:"env:KUBERNETES_NAMESPACE"`
}

type PingConfig struct {
	PingValue string `conf:"env:PING_VALUE, default:ping"`
}

// Filepath is used as a custom decoder which will take a configuration string
// and resolve a ~ to the absolute path of the home directory. If ~ is not
// present it treated as a normal path to a directory
type Filepath struct {
	Path string
}

func (d *Filepath) String() string {
	return d.Path
}

func (d *Filepath) IsEmpty() bool {
	return d.Path == ""
}

func (d *Filepath) Decode(v string) error {
	if v == "" {
		return nil
	}

	path, err := homedir.Expand(v)
	if err != nil {
		return failure.ToInvalidParam(err, "homedir.Expand failed (%s)", v)
	}

	d.Path = path
	return nil
}
