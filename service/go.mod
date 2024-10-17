module github.com/pennsieve/processor-pre-metadata/service

go 1.21

replace github.com/pennsieve/processor-pre-metadata/client => ./../client

require (
	github.com/google/uuid v1.6.0
	github.com/pennsieve/processor-pre-metadata/client v0.0.0-20241015190912-3e004748369b
	github.com/stretchr/testify v1.9.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
