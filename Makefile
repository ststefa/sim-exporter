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

help:   ## Show this help
	@fgrep -h "##" $(MAKEFILE_LIST) | fgrep -v fgrep | sed 's/:.*##/:/'


local_build: _build ## Build executable locally (mainly for development purpose)

local_image:   ## Create docker image locally (mainly for development purpose)
	docker build -f local.Dockerfile --tag $(EXEC_NAME) --progress=plain .

local_image_test:  ## Test the docker image
	docker run --rm $(EXEC_NAME)

local_clean: ## Remove build output
	rm -vfr build

alm_build: _build ## Build executable in GEC ALM CI pipeline

_build:
# The ALM process _requires_ all output to be located under the build/ dir
	mkdir -pv build
	# CGO_ENABLED: required for "docker FROM SCRATCH"
	# -w: Omit DWARF information
	# -X: Override variables at link-time
	CGO_ENABLED=0 go build -ldflags "-w" -ldflags "-s" -ldflags "-X '$(MOD_NAME)/cmd.version=$(VERSION)'" -o build/$(EXEC_NAME)
	# Compress binary
	# Produces invalid executables on aarch64, skipping
	#upx --no-progress build/$(EXEC_NAME)
