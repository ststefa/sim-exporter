# Buildfile for both
# - Local development (targets "local...")
# - ALM pipeline (targets "alm...")

.DEFAULT_GOAL := help

# Get version from auto-bumped gitlab ALM version.txt file
VERSION := $(shell cat version.txt)

# Get our own module name.
MOD_NAME := $(shell go list -m)

# Name of resulting executable (last element of module name by convention)
EXEC_NAME := $(shell basename ${MOD_NAME})

# Docker hub user, only required for my_multiarch_image
ME := ststefa

help:   ## Show this help
	@grep -h "##" $(MAKEFILE_LIST) | grep -v grep | sed 's/:.*##/:/'

local_build: _check_netrc _build ## Create build artifacts in build/ directory

local_test:  ## Run tests
	go test ./...

local_image: _check_netrc   ## Create docker image locally (mainly for development purpose)
	docker build -f local.Dockerfile --secret id=netrc,src=${HOME}/.netrc --tag $(EXEC_NAME) --progress=plain .

my_multiarch_image: _check_netrc   ## Create multi-arch docker image and push it to dockerhub (dev workaround, only usable for $ME). Note that multi-arch builds require additional docker setup!
	docker buildx build --secret id=netrc,src=${HOME}/.netrc --platform linux/amd64,linux/arm64,linux/arm/v7 -t $(ME)/$(EXEC_NAME) --push -f local.Dockerfile --progress=plain .

local_image_test:  ## Test the docker image
	docker run --rm $(EXEC_NAME)

local_clean: ## Remove build directory
	rm -vfr build

alm_build: _build ## Build executable in GEC ALM CI pipeline

# The remaining targets are "private" and not meant to be invoked directly.
# This cannot be technically enforced but is indicated with the "_" prefix.

_check_netrc:
	@[ -f ~/.netrc ] || (echo "~/.netrc is required to access non-public go repo. See https://git.mgmt.innovo-cloud.de/operations-center/operationscenter-observability/observability-cli/-/blob/master/README.md#golang-and-private-repos." ; false)

_build:
# The ALM process _requires_ all output to be located under the build/ dir
	mkdir -pv build
	# CGO_ENABLED: required for "docker FROM SCRATCH"
	# -w: Omit DWARF information
	# -X: Override variables at link-time
	CGO_ENABLED=0 go build -ldflags "-w" -ldflags "-s" -ldflags "-X '$(MOD_NAME)/cmd.version=$(VERSION)'" -o build/$(EXEC_NAME)
	cp -pr examples build
