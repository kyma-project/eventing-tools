#syntax=docker/dockerfile-upstream:1.4
FROM eu.gcr.io/kyma-project/external/golang:1.20.1-alpine3.17 as builder

WORKDIR /app

ENV GOPATH /go

COPY . /app

# Build
RUN go install ./cmd/loadtest-subscriber && \
  go install ./cmd/loadtest-publisher && \
  go install ./cmd/publisher && \
  go install ./cmd/subscriber

FROM gcr.io/distroless/static:nonroot
LABEL source = git@github.com:kyma-project/kyma.git

WORKDIR /
COPY --from=builder /go/bin/* /usr/local/bin/
USER nonroot:nonroot
