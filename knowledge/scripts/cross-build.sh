#!/bin/bash
set -ex

LD_FLAGS="-s -w -X github.com/gptscript-ai/knowledge/version.Version=${GIT_TAG}"

#
# Main build - includes MuPDF, which requires CGO and is currently not possible to be built for linux/arm64 and windows/arm64
#

export CGO_ENABLED=1
if [ "$(go env GOOS)" = "linux" ]; then
  # Linux: amd64, arm64
  GOARCH=amd64 go build -o dist/knowledge-linux-amd64 -ldflags "${LD_FLAGS}" .
else

  # Windows: amd64
  CC=x86_64-w64-mingw32-gcc GOOS=windows GOARCH=amd64 go build -o dist/knowledge-windows-amd64.exe -tags "${GO_TAGS}" -ldflags "${LD_FLAGS}" .

  # Darwin: amd64, arm64
  GOARCH=amd64 go build -o dist/knowledge-darwin-amd64 -ldflags "${LD_FLAGS}" .
  GOARCH=arm64 go build -o dist/knowledge-darwin-arm64 -ldflags "${LD_FLAGS}" .
fi

#
# NO CGO build - Does NOT include MuPDF, which requires CGO and is currently not possible to be built for linux/arm64 and windows/arm64
#
if [ "$(go env GOOS)" = "linux" ]; then
  GO_TAGS="netgo"
  export CGO_ENABLED=0
  # Linux: arm64
  GOARCH=arm64 go build -o dist/knowledge-linux-arm64 -tags "${GO_TAGS}" -ldflags "${LD_FLAGS}\" -extldflags \"-static\" " .

  # Windows: arm64
  GOARCH=arm64 GOOS=windows go build -o dist/knowledge-windows-arm64.exe -tags "${GO_TAGS}" -ldflags "${LD_FLAGS}" .
fi
