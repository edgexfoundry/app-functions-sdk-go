#
# Copyright (c) 2021 Intel Corporation
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

.PHONY: build test clean docker

# change the following boolean flag to enable or disable the Full RELRO (RELocation Read Only) for linux ELF (Executable and Linkable Format) binaries
ENABLE_FULL_RELRO:="true"
# change the following boolean flag to enable or disable PIE for linux binaries which is needed for ASLR (Address Space Layout Randomization) on Linux, the ASLR support on Windows is enabled by default
ENABLE_PIE:="true"

GOTESTFLAGS?=-race

# VERSION file is not needed for local development, In the CI/CD pipeline, a temporary VERSION file is written
# if you need a specific version, just override below
# TODO: If your service is not being upstreamed to Edgex Foundry, you need to determine the best approach for
#       setting your service's version for non-development builds.
APPVERSION=$(shell cat ./VERSION 2>/dev/null || echo 0.0.0)

# This pulls the version of the SDK from the go.mod file. If the SDK is the only required module,
# it must first remove the word 'required' so the offset of $2 is the same if there are multiple required modules
SDKVERSION=$(shell cat ./go.mod | grep 'github.com/edgexfoundry/app-functions-sdk-go/v3 v' | sed 's/require//g' | awk '{print $$2}')

MICROSERVICE=app-new-service
GOFLAGS=-ldflags "-s -w -X github.com/edgexfoundry/app-functions-sdk-go/v4/internal.SDKVersion=$(SDKVERSION) -X github.com/edgexfoundry/app-functions-sdk-go/v4/internal.ApplicationVersion=$(APPVERSION)" -trimpath -mod=readonly

# TODO: uncomment and remove default once files are in a Github repository or
#       remove totally including usage below
#GIT_SHA=$(shell git rev-parse HEAD)
GIT_SHA=no-sha

ifeq ($(ENABLE_FULL_RELRO), "true")
	GOFLAGS += -ldflags "-bindnow"
endif

ifeq ($(ENABLE_PIE), "true")
	GOFLAGS += -buildmode=pie
endif

# CGO is enabled by default and causes docker builds to fail due to no gcc,
# but is required for test with -race, so must disable it for the builds only
build: tidy
	CGO_ENABLED=0 go build $(GOFLAGS) -o $(MICROSERVICE)

tidy:
	go mod tidy

# TODO: Change the registries (edgexfoundry, nexus3.edgexfoundry.org:10004) below as needed.
#       Leave them as is if service is to be upstreamed to EdgeX Foundry
# NOTE: This is only used for local development. Jenkins CI does not use this make target
docker:
	docker build \
	    --build-arg http_proxy \
	    --build-arg https_proxy \
		-f Dockerfile \
		--label "git_sha=$(GIT_SHA)" \
		-t edgexfoundry/app-new-service:$(GIT_SHA) \
		-t edgexfoundry/app-new-service:${APPVERSION}-dev \
		-t nexus3.edgexfoundry.org:10004/app-new-service:${APPVERSION}-dev \
		.

# The test-attribution-txt.sh scripts are required for upstreaming to EdgeX Foundry.
# TODO: Remove bin folder and reference to script below if NOT upstreaming to EdgeX Foundry.
test:
	go test $(GOTESTFLAGS) -coverprofile=coverage.out ./...
	go vet ./...
	gofmt -l $$(find . -type f -name '*.go'| grep -v "/vendor/")
	[ "`gofmt -l $$(find . -type f -name '*.go'| grep -v "/vendor/")`" = "" ]
	./bin/test-attribution-txt.sh

clean:
	rm -f $(MICROSERVICE)

vendor:
	go mod vendor
