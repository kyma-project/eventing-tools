#syntax=docker/dockerfile-upstream:1.4
FROM eu.gcr.io/kyma-project/external/golang:1.20.3-alpine3.17 as builder

WORKDIR /app

ENV GOPATH /go

COPY . /app

# Build
RUN go install . 

FROM gcr.io/distroless/static:nonroot
LABEL source = git@github.com:kyma-project/eventing-tools.git

WORKDIR /
COPY --from=builder /go/bin/* /usr/local/bin/
USER nonroot:nonroot
ENTRYPOINT ["/usr/local/bin/eventing-tools"]
