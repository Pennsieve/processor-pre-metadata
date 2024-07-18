# Metadata Pre-Processor

Retrieves a dataset's metadata and places it in the input directory.

To build:

`docker build -t pennsieve/metadata-pre-processor .`

On arm64 architectures:

`docker build -f Dockerfile_arm64 -t pennsieve/metadata-pre-processor .`

To run tests:

` go test ./...`

To run integration test:

1. Given a dataset you want to test with, create an integration for the dataset and this pre-processor. Get the
   integration id
2. Copy `dev.env.example` to `dev.env`
3. In `dev.env` update `SESSION_TOKEN` with a valid token and `INTEGRATION_ID` with the id from the first step.
4. Run `./run-integration-test.sh dev.env`

