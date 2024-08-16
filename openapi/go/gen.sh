#!/bin/sh

# NOTE: This script assumes it is run from project root

mkdir -p openapi/go/service

go run github.com/deepmap/oapi-codegen/cmd/oapi-codegen@v1.16.2 -config openapi/go/oapi-codegen.yml openapi/openapi.yml

go mod tidy

go run github.com/vektra/mockery/v2@v2.43.0
