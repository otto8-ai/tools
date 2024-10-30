# get git tag
ifneq ($(GIT_TAG_OVERRIDE),)
$(info GIT_TAG set from env override!)
GIT_TAG := $(GIT_TAG_OVERRIDE)
endif

GIT_TAG   ?= $(shell git describe --tags)
ifeq ($(GIT_TAG),)
GIT_TAG   := $(shell git describe --always)
endif

GO_TAGS := netgo
LD_FLAGS := -s -w -X github.com/gptscript-ai/knowledge/version.Version=${GIT_TAG}
build:
	go build -mod=mod -o bin/knowledge -tags "${GO_TAGS}" -ldflags '$(LD_FLAGS) ' .

run: build
	bin/knowledge server

run-dev: generate run

clean-dev:
	rm -rf knowledge.db vector.db

generate: tools openapi

openapi:
	swag init --parseDependency -g pkg/server/server.go -o pkg/docs

tools:
	if ! command -v swag &> /dev/null; then go install github.com/swaggo/swag/cmd/swag@latest; fi

lint:
	golangci-lint run ./...

test:
	go test -v ./...

build-cross:
	GIT_TAG=${GIT_TAG} ./scripts/cross-build.sh

ci-setup:
	@echo "### Installing Go tools..."
	@echo "### -> Installing golangci-lint..."
	curl -sfL $(PKG_GOLANGCI_LINT_SCRIPT) | sh -s -- -b $(GOENVPATH)/bin v$(PKG_GOLANGCI_LINT_VERSION)

cli-docs:
	go run -mod=mod docs/gendocs/main.go

serve-docs:
	(cd docs && yarn install && yarn start)

build-docs:
	(cd docs && yarn install --frozen-lockfile --non-interactive && yarn build)

# This will initialize the node_modules needed to run the docs dev server. Run this before running serve-docs
init-docs:
	docker run --rm --workdir=/docs -v $${PWD}/docs:/docs node:18-buster yarn install

# Ensure docs build without errors. Makes sure generated docs are in-sync with CLI.
validate-docs:
	docker run --rm --workdir=/docs -v $${PWD}/docs:/docs node:18-buster yarn build
	go run tools/gendocs/main.go
	if [ -n "$$(git status --porcelain --untracked-files=no)" ]; then \
		git status --porcelain --untracked-files=no; \
		echo "Encountered dirty repo!"; \
		git diff; \
		exit 1 \
	;fi

gen-checksum:
	$$(cd dist && sha256sum knowledge-* > checksums.txt)