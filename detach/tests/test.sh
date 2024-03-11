#!/usr/bin/env bash

# start main go program
go run detach/tests/main.go

# detached/orphane code `detachedFunc()` is still running and will write "/tmp/detach.test.txt"

sleep 10

if [ -f "/tmp/detach.test.txt" ]; then
    echo "ok"
    exit 0
else
    echo "ko"
    exit 1
fi