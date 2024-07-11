# Metadata Pre-Processor

Retrieves a dataset's metadata and places it in the input directory.

To build:

`docker build -t pennsieve/metadata-pre-processor .`

On arm64 architectures:

`docker build -f Dockerfile_arm64 -t pennsieve/metadata-pre-processor .`

To run:

`docker-compose up --build`

To run tests:

` go test ./...`

