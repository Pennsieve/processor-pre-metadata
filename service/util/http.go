package util

import (
	"github.com/pennsieve/processor-pre-metadata/service/logging"
	"log/slog"
	"net/http"
)

var logger = logging.PackageLogger("util")

func CloseAndWarn(response *http.Response) {
	if err := response.Body.Close(); err != nil {
		logger.Warn("error closing response body",
			slog.String("method", response.Request.Method),
			slog.String("url", response.Request.URL.String()),
			slog.Any("error", err))
	}
}
