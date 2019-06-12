#!/usr/bin/env bash

FILES=$(go list ./... | egrep -v "gocat|vendor")
set -e
echo "" > coverage.txt

go version 

for d in $FILES; do
    go test -coverprofile=profile.out -covermode=atomic $d
    if [ -f profile.out ]; then
        cat profile.out >> coverage.txt
        rm profile.out
    fi
done
