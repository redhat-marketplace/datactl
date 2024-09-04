FROM registry.access.redhat.com/ubi9/go-toolset:1.21 AS build
ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG TARGETOS
ARG TARGETARCH
ARG FIPS_DETECT_VERSION=7157dae

# Binary and pkg destination
RUN mkdir -p /opt/app-root/src/go/bin && \
    mkdir -p /opt/app-root/src/go/pkg/

# HOME=/opt/app-root/src/
# Mount code, build cache, mod cache
# cache id must be set to get desired uid for mount
RUN --mount=type=bind,source=.,readonly,target=/opt/app-root/src/go/src/github.com/redhat-marketplace/datactl \
    --mount=type=cache,id=go-build,uid=1001,gid=0,target=/opt/app-root/src/.cache/go-build \
    --mount=type=cache,id=mod,uid=1001,gid=0,target=/opt/app-root/src/go/pkg/mod \
    cd /opt/app-root/src/go/src/github.com/redhat-marketplace/datactl && \
    go version && \
    go mod download && \
    GOFLAGS="-buildvcs=false" go install ./cmd/datactl && \
    go install github.com/acardace/fips-detect@${FIPS_DETECT_VERSION}


FROM registry.access.redhat.com/ubi9/ubi-minimal
ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG TARGETOS
ARG TARGETARCH

COPY --from=build /opt/app-root/src/go/bin/datactl /usr/local/bin/datactl
COPY --from=build /opt/app-root/src/go/bin/fips-detect /usr/local/bin/fips-detect
COPY entrypoint.sh .

ENV OPENSSL_FORCE_FIPS_MODE=1

ENTRYPOINT ["./entrypoint.sh"]
