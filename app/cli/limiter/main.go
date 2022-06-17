// Package main is the entry point to cli system. This is how we start, stop
// and debug our http service
package main

import (
	"github.com/rsb/api_rate_limiter/app/cli/limiter/cmd"
)

var build = "develop"

func main() {
	cmd.Execute(build)
}
