package main

import (
	"github.com/pennsieve/processor-pre-metadata/logging"
	"github.com/pennsieve/processor-pre-metadata/preprocessor"
	"log/slog"
	"os"
)

var logger = logging.PackageLogger("main")

func main() {

	m := preprocessor.FromEnv()

	logger.Info("created MetadataPreProcessor",
		slog.String("integrationID", m.IntegrationID),
		slog.String("baseDirectory", m.BaseDirectory),
		slog.String("APIHost", m.Pennsieve.APIHost),
		slog.String("API2Host", m.Pennsieve.API2Host),
	)

	if err := m.Run(1000, 1000); err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}
