package util

import (
	"github.com/pennsieve/processor-pre-metadata/service/logging"
	"net/http"
)

var logger = logging.PackageLogger("util")

func CloseAndWarn(response *http.Response) {
	if err := response.Body.Close(); err != nil {
		logger.Warn("error closing response body from %s %s: %w", response.Request.Method, response.Request.URL, err)
	}
}
