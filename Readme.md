# Metadata Pre-Processor

Retrieves a dataset's metadata and places it in the input directory.

To build:

`docker build -t pennsieve/metadata-pre-processor .`

On arm64 architectures:

`docker build -f Dockerfile_arm64 -t pennsieve/metadata-pre-processor .`

To run tests:

` go test ./...`

To run integration test:

1. Copy `dev.env.example` to `dev.env`
2. In `dev.env` update `SESSION_TOKEN` with a valid token
3. Run `./run-integration-test.sh dev.env`

