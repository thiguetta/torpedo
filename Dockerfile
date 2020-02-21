FROM golang:1.13.7-alpine AS build
LABEL maintainer="harsh@portworx.com"

WORKDIR /go/src/github.com/portworx/torpedo

# Install setup dependencies
RUN apk update && \
    apk add git gcc  musl-dev && \
    apk add make && \
    apk add openssh-client && \
    apk add curl

# No need to copy *everything*. This keeps the cache useful
COPY deployments deployments
COPY cmd cmd
COPY drivers drivers
COPY pkg pkg
COPY scripts scripts
COPY tests tests
COPY vendor vendor
COPY Makefile Makefile
COPY go.mod go.mod
COPY go.sum go.sum
COPY main.go main.go

# Why? Errors if this is removed
COPY .git .git

# Compile
RUN mkdir bin && \
    make build && \
    make install

RUN curl -fsSL -O https://get.helm.sh/helm-v3.1.1-linux-amd64.tar.gz && \
    tar -zxvf helm-v3.1.1-linux-amd64.tar.gz && \
    mv linux-amd64/helm /bin/helm

RUN git clone https://github.com/pingcap/chaos-mesh.git

# Build a fresh container with just the binaries
FROM alpine

RUN apk add ca-certificates 

# Install kubectl from Docker Hub.
COPY --from=lachlanevenson/k8s-kubectl:latest /usr/local/bin/kubectl /usr/local/bin/kubectl

# Copy scripts into container
WORKDIR /torpedo
COPY deployments deployments
COPY scripts scripts

WORKDIR /go/src/github.com/portworx/torpedo

# Copy ginkgo & binaries over from previous container
COPY --from=build /bin/helm /bin/helm
COPY --from=build /go/bin/ginkgo /bin/ginkgo
COPY --from=build /go/bin/torpedo /bin/torpedo
COPY --from=build /go/src/github.com/portworx/torpedo/bin bin
COPY --from=build /go/src/github.com/portworx/torpedo/chaos-mesh/manifests/ manifests
COPY --from=build /go/src/github.com/portworx/torpedo/chaos-mesh/helm/chaos-mesh chaos-mesh
COPY drivers drivers

ENTRYPOINT ["torpedo"]
CMD []
