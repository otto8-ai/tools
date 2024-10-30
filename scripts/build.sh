#!/bin/bash
set -e

cd $(dirname $0)/..

for gomod in $(find . -name go.mod); do
    if [ $(basename $(dirname $gomod)) == common ]; then
        continue
    fi
    (
        cd $(dirname $gomod)
        echo Building $PWD
        go build -ldflags="-s -w" -o bin/gptscript-go-tool .
    )
done

