#!/bin/bash
set -ex

LD_FLAGS="-s -w -X github.com/gptscript-ai/knowledge/version.Version=${GIT_TAG}"

#
# Main build - includes MuPDF, which requires CGO and is currently not possible to be built for linux/arm64 and windows/arm64
#

export CGO_ENABLED=1
if [ "$(go env GOOS)" = "linux" ]; then
  # Linux: amd64
  GOARCH=amd64 go build -o dist/knowledge-linux-amd64 -ldflags "${LD_FLAGS}" .

  if [ "$(go env GOARCH)" = "amd64" ]; then
    # Linux: arm64 (on amd64) - apt install gcc-aarch64-linux-gnu
    CC=aarch64-linux-gnu-gcc GOARCH=arm64 go build -o dist/knowledge-linux-arm64 -ldflags "${LD_FLAGS}" .
  elif [ "$(go env GOARCH)" = "arm64" ]; then
    # Linux: arm64 (on arm64)
    GOARCH=arm64 go build -o dist/knowledge-linux-arm64 -tags "${GO_TAGS}" -ldflags "${LD_FLAGS}" .
  fi
else

  # This is expected to run on a MacOS/Darwin machine with mingw32 installed for the Windows builds

  # Windows: amd64
  CC=x86_64-w64-mingw32-gcc GOOS=windows GOARCH=amd64 go build -o dist/knowledge-windows-amd64.exe -tags "${GO_TAGS}" -ldflags "${LD_FLAGS}" .

  # Windows: arm64 -> TODO: currently disabled as we need a special compiler from the llvm-mingw project for this
  # GOARCH=arm64 GOOS=windows go build -o dist/knowledge-windows-arm64.exe -tags "${GO_TAGS}" -ldflags "${LD_FLAGS}" .

  # Darwin: amd64, arm64
  GOARCH=amd64 go build -o dist/knowledge-darwin-amd64 -ldflags "${LD_FLAGS}" .
  GOARCH=arm64 go build -o dist/knowledge-darwin-arm64 -ldflags "${LD_FLAGS}" .
fi