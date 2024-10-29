#!/bin/sh

# initArch discovers the architecture for this system.
initArch() {
  if [ -z $ARCH  ]; then
    ARCH=$(uname -m)
    case $ARCH in
      armv5*) ARCH="armv5";;
      armv6*) ARCH="armv6";;
      armv7*) ARCH="arm";;
      aarch64) ARCH="arm64";;
      x86) ARCH="386";;
      x86_64) ARCH="amd64";;
      i686) ARCH="386";;
      i386) ARCH="386";;
    esac
  fi
}

# initOS discovers the operating system for this system.
initOS() {
  if [ -z $OS ]; then
    OS=$(uname|tr '[:upper:]' '[:lower:]')

    case "$OS" in
      # Minimalist GNU for Windows
      mingw*) OS='windows';;
    esac
  fi
}


install_golangci_lint() {
  echo "Installing golangci-lint..."
  GOLANGCI_LINT_VERSION=v1.57.2
  curl -sSfL "https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh" | sh -s -- -b $(go env GOPATH)/bin ${GOLANGCI_LINT_VERSION}
}

install_gox() {
  echo "Installing gox for $OS/$ARCH..."
  GOX_REPO=iwilltry42/gox
  GOX_VERSION=0.1.0
  curl -sSfL "https://github.com/${GOX_REPO}/releases/download/v${GOX_VERSION}/gox_${GOX_VERSION}_${OS}_${ARCH}.tar.gz" | tar -xz -C /tmp
  chmod +x /tmp/gox
  mv /tmp/gox /usr/local/bin/gox
}

#
# MAIN
#

initOS
initArch

for pkg in "$@"; do
  case "$pkg" in
    golangci-lint) install_golangci_lint;;
    gox) install_gox;;
    *) printf "ERROR: Unknown Package '%s'" $pkg;;
  esac
done
