#!/bin/bash
[ -L app/docs ] || [ -d app/docs ] || ln -s ../docs app/docs
cd ./app && go test ./... -v -coverprofile=coverage.out
go tool cover -func=coverage.out