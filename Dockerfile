FROM golang:1.9.2-alpine AS build
LABEL maintainer="harsh@portworx.com"

WORKDIR /go/src/github.com/portworx/torpedo

# Install setup dependencies
RUN apk update && \
    apk add git && \
    apk add curl && \
    apk add make && \
    apk add openssh-client && \
    go get github.com/onsi/ginkgo/ginkgo && \
    go get github.com/onsi/gomega && \
    go get github.com/sirupsen/logrus && \
    go get -u github.com/go-delve/delve/cmd/dlv

# No need to copy *everything*. This keeps the cache useful
COPY deployments deployments
COPY drivers drivers
COPY pkg pkg
COPY scripts scripts
COPY tests tests
COPY vendor vendor
COPY Makefile Makefile

# Why? Errors if this is removed
COPY .git .git

# Compile
RUN mkdir bin && \
    make build

# Build a fresh container with just the binaries
FROM alpine

# Copy scripts into container
WORKDIR /torpedo
COPY deployments deployments
COPY scripts scripts

WORKDIR /go/src/github.com/portworx/torpedo

# Copy ginkgo & binaries over from previous container
COPY --from=build /go/bin/ginkgo /bin/ginkgo
COPY --from=build /go/bin/dlv /go/bin/dlv
COPY --from=build /go/src/github.com/portworx/torpedo/bin bin
COPY drivers drivers

ENTRYPOINT ["dlv", "--listen=:2345", "--headless=true", "--api-version=2", "--accept-multiclient", "exec", "ginkgo", "--failFast", "--slowSpecThreshold", "180", "-v", "-trace"]
CMD []

EXPOSE 2345
