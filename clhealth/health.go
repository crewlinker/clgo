// Package clhealth provides healthcheck functionality.
package clhealth

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/samber/lo"
)

// how long the check might take.
var (
	healthCheckTimeout = time.Second * 2
	invalidExitCode    = 2
)

// CheckHealthAndExit will run the fx.App unless the osArgs indicate that the
// user wants to run a healthcheck. This is useful since our containers might
// not contain curl or wget.
func CheckHealthAndExit(ctx context.Context, errWriter io.Writer, osArgs []string, exitFn func(int)) {
	if len(osArgs) != 4 || osArgs[1] != "healthcheck" {
		return
	}

	// setup timeout
	ctx, cancel := context.WithTimeout(ctx, healthCheckTimeout)
	defer cancel()

	// init request
	req := lo.Must(http.NewRequestWithContext(ctx, http.MethodGet, osArgs[2], nil))

	// perform request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Fprintf(errWriter, "%v", err)
		exitFn(invalidExitCode)

		return
	}

	resp.Body.Close()

	// if we have an unexpected status code, exit with code 1
	if strconv.Itoa(resp.StatusCode) != osArgs[3] {
		fmt.Fprintf(errWriter, "%v", resp.Status)
		exitFn(1)

		return
	}

	// else, exit with code 0
	exitFn(0)
}
