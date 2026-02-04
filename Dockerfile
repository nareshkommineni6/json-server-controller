# Build the manager binary
FROM golang:1.21 AS builder
ARG TARGETOS
ARG TARGETARCH

WORKDIR /workspace

# Copy everything first
COPY go.mod go.mod
COPY go.sum* ./
COPY cmd/ cmd/
COPY api/ api/
COPY internal/ internal/

# Download dependencies (go mod tidy needs source files to know what to fetch)
RUN go mod tidy && go mod download

# Build
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} go build -a -o manager cmd/main.go

# Use distroless as minimal base image to package the manager binary
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workspace/manager .
USER 65532:65532

ENTRYPOINT ["/manager"]
