# Build the manager binary
FROM golang:1.15 as builder

LABEL name="JanusGraph Operator Using Cassandra" \
  vendor="IBM" \
  version="v0.0.1" \
  release="1" \
  summary="This is a JanusGraph operator that ensures stateful deployment in an OpenShift cluster." \
  description="This operator will deploy JanusGraph in OpenShift cluster."

  # Required Licenses for Red Hat build service and scanner
COPY licenses /licenses

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY main.go main.go
COPY api/ api/
COPY controllers/ controllers/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o manager main.go

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workspace/manager .

USER 65532:65532

ENTRYPOINT ["/manager"]
