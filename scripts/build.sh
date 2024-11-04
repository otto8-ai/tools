#!/bin/bash
set -e

cd $(dirname $0)/..

for maingo in $(find -L . -name main.go); do
    if [ $(basename $(dirname $maingo)) == common ]; then
        continue
    fi
    (
        cd $(dirname $maingo)
        echo Building $PWD
        go build -ldflags="-s -w" -o bin/gptscript-go-tool .
    )
done
