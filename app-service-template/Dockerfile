# TODO: Change Copyright to your company if open sourcing or remove header
#
# Copyright (c) 2022 Intel Corporation
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

#build stage
ARG BASE=golang:1.23-alpine3.20
FROM ${BASE} AS builder

ARG ALPINE_PKG_BASE="make git"
ARG ALPINE_PKG_EXTRA=""

RUN apk add --update --no-cache ${ALPINE_PKG_BASE} ${ALPINE_PKG_EXTRA}
WORKDIR /app

COPY go.mod vendor* ./
RUN [ ! -d "vendor" ] && go mod download all || echo "skipping..."

COPY . .
ARG MAKE="make build"
RUN $MAKE

#final stage
FROM alpine:3.20
# TODO: Change Copyright to your company if open sourcing or remove label
LABEL license='SPDX-License-Identifier: Apache-2.0' \
  copyright='Copyright (c) 2022: Intel'
LABEL Name=app-new-service Version=${VERSION}

# dumb-init is required as security-bootstrapper uses it in the entrypoint script
RUN apk add --update --no-cache ca-certificates dumb-init
# Ensure using latest versions of all installed packages to avoid any recent CVEs
RUN apk --no-cache upgrade

COPY --from=builder /app/Attribution.txt /Attribution.txt
COPY --from=builder /app/LICENSE /LICENSE
COPY --from=builder /app/res/ /res/
COPY --from=builder /app/app-new-service /app-new-service

# TODO: set this port appropriatly as it is in the configuation.yaml
EXPOSE 59740

ENTRYPOINT ["/app-new-service"]
CMD ["-cp=keeper.http://edgex-core-keeper:59890", "--registry"]
