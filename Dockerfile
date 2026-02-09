#syntax=docker/dockerfile-upstream:1.4
FROM europe-docker.pkg.dev/kyma-project/prod/external/golang:1.22.2-alpine3.19 as builder

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
