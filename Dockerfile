FROM registry.access.redhat.com/ubi8/go-toolset AS build

# Binary destination
RUN mkdir -p /opt/app-root/src/go/bin

# HOME=/opt/app-root/src/
# Mount code, build cache, mod cache
# cache id must be set to get desired uid for mount
RUN --mount=type=bind,source=.,readonly,target=/opt/app-root/src/go/src/github.com/redhat-marketplace/datactl \
    --mount=type=cache,id=go-build,uid=1001,gid=0,target=/opt/app-root/src/.cache/go-build \
    --mount=type=cache,id=mod,uid=1001,gid=0,target=/opt/app-root/src/go/pkg/mod \
    cd /opt/app-root/src/go/src/github.com/redhat-marketplace/datactl && \
    go mod download && \
    GOFLAGS="-buildvcs=false" go install ./cmd/datactl

FROM registry.access.redhat.com/ubi8/ubi-micro
COPY --from=build /opt/app-root/src/go/bin/datactl .
ENV OPENSSL_FORCE_FIPS_MODE=1
CMD ./datactl
