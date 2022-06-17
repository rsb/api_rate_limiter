// Package construct is used to build (construct) dependencies to be injected
// into api handlers or cli entry points
package construct

import (
	"github.com/rsb/api_rate_limiter/app"
	"github.com/rsb/api_rate_limiter/app/api/handlers/health"
	"github.com/rsb/api_rate_limiter/app/conf"
	"github.com/rsb/api_rate_limiter/foundation/logging"
	"github.com/rsb/failure"
	"go.uber.org/zap"
	"net/http"
	"time"

	"github.com/gofiber/contrib/fiberzap"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	expvarmw "github.com/gofiber/fiber/v2/middleware/expvar"
	"github.com/gofiber/fiber/v2/middleware/pprof"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

const (
	DefaultHTTPClientTimeout             = 5 * time.Second
	DefaultHTTPClientMaxIde              = 100
	DefaultHTTPClientMaxConnsPerHost     = 100
	DefaultHTTPClientMaxIdleConnsPerHost = 100
)

func NewLogger(appVersion string) (*zap.SugaredLogger, error) {
	l, err := logging.NewLogger(app.ServiceName, appVersion)
	if err != nil {
		return nil, failure.Wrap(err, "logging.NewLogger failed")
	}

	return l, nil
}

func NewHttpClient(config conf.HTTPClient) *http.Client {
	t := http.DefaultTransport.(*http.Transport).Clone()
	t.MaxIdleConns = config.MaxIdleConn
	t.MaxConnsPerHost = config.MaxConnPerHost
	t.MaxIdleConnsPerHost = config.MaxIdleConnPerHost

	return &http.Client{
		Timeout:   config.Timeout,
		Transport: t,
	}
}

func NewAPIMux(d app.Dependencies, c conf.API) *fiber.App {

	app := fiber.New(c.NewFiberConfig())
	app.Use(recover.New())
	app.Use(cors.New())
	app.Use(fiberzap.New(
		fiberzap.Config{
			Logger: d.Logger.Desugar(),
		},
	))

	return app
}

func NewDefaultHTTPClient() *http.Client {
	config := conf.HTTPClient{
		Timeout:            DefaultHTTPClientTimeout,
		MaxIdleConn:        DefaultHTTPClientMaxIde,
		MaxConnPerHost:     DefaultHTTPClientMaxConnsPerHost,
		MaxIdleConnPerHost: DefaultHTTPClientMaxIdleConnsPerHost,
	}

	return NewHttpClient(config)
}

// NewDebugMux registers all the debug standard library routes and then custom
// debug application routes for the service. This bypassing the use of the
// DefaultServerMux. Using the DefaultServerMux would be a security risk since
// a dependency could inject a handler into our service without us knowing it.
func NewDebugMux(d *app.Dependencies) *fiber.App {
	r := fiber.New()
	r.Use(pprof.New())
	r.Use(expvarmw.New())
	h := health.NewCheckHandler(d)

	r.Get("/debug/readiness", h.Readiness)
	r.Get("/debug/liveness", h.Liveness)

	return r
}