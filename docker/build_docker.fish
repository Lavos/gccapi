#!/usr/bin/fish

set -x CGO_ENABLED 0
go build -a -installsuffix cgo -o service ../main.go
