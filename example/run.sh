#!/bin/sh

go build github.com/cbarrick/evo/...

exec go run $GOPATH/src/github.com/cbarrick/evo/example/$1/main.go $@
