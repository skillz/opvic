# Build the agent binary
FROM golang:1.16 as builder

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY cmd/controlplane .
COPY utils/ utils/
COPY controlplane/ controlplane/
COPY agent/api agent/api

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o opvic main.go

# Use distroless as minimal base image to package the agent binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workspace/opvic .
USER 65532:65532

ENTRYPOINT ["/opvic"]
