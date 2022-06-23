// Package tests is used model and run integration tests that
// are used to ensure we can actually limit the rate of requests
package tests

import (
	"github.com/gofiber/fiber/v2"
	"github.com/rsb/api_rate_limiter/app"
	"github.com/rsb/api_rate_limiter/app/conf"
	"github.com/rsb/api_rate_limiter/app/construct"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

func NewAPI(t *testing.T, config conf.API) (*fiber.App, app.Dependencies) {
	logger, err := construct.NewLogger("testing")
	require.NoError(t, err, "construct.NewLogger should not failed")

	depend := app.Dependencies{
		Logger: logger,
	}

	app := construct.NewAPIMux(config, logger)
	app = construct.AddAllRoutes(app, &depend)

	return app, depend
}

func TestRateLimiting(t *testing.T) {
	config := conf.API{
		RateLimit:         50,
		RateLimitInterval: 2 * time.Second,
	}

	app, _ := NewAPI(t, config)

	var wg sync.WaitGroup
	singleRequest := func(wg *sync.WaitGroup) {
		defer wg.Done()

		resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/ping", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Equal(t, `{"value":"pong"}`, string(body))
	}

	for i := 0; i <= 49; i++ {
		wg.Add(1)
		go singleRequest(&wg)
	}

	wg.Wait()

	resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/ping", nil))
	require.NoError(t, err)
	require.Equal(t, http.StatusTooManyRequests, resp.StatusCode)

	time.Sleep(3 * time.Second)

	resp, err = app.Test(httptest.NewRequest(http.MethodGet, "/ping", nil))
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
}
