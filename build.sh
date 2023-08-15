#!/bin/env bash

go mod tidy

go test -v ./... && go build