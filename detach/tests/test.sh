#!/usr/bin/env bash

# start main go program
pwd
go run detach/tests/main.go

# detached/orphane code `detachedFunc()` is still running and will write "/tmp/detach.test.txt"

sleep 10

if [ -f "/tmp/detach.test.txt" ]; then
    echo "ko"
    exit 0
else
    echo "ok"
    exit 1
fi