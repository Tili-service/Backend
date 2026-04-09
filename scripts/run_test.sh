#!/bin/bash
cd ./app && go test ./... -v -coverprofile=coverage.out
go tool cover -func=coverage.out