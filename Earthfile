VERSION 0.6

ARG GO_VERSION=1.17

FROM golang:${GO_VERSION}

WORKDIR /build

all:
    BUILD +build
    BUILD +test
    BUILD +gen-proto
    # Enable on next PR
    #BUILD +breaking-proto

GO_RUN:
    COMMAND
    ARG --required cmd
    ARG GOCACHE=/go-cache
    RUN --mount=type=cache,target=$GOCACHE $cmd

deps:
    COPY go.mod go.sum .
    DO +GO_RUN --cmd="go mod download all"

build:
    FROM +deps
    COPY *.go .
    DO +GO_RUN --cmd="go build ./..."

test:
    FROM +deps
    COPY *.go .
    DO +GO_RUN --cmd="go test -v ./..."

proto-deps:
    ARG TARGETOS
    ARG TARGETARCH
    IF [ "${TARGETARCH}" = "arm64" ]
        ARG PROTOARCH=aarch_64
    ELSE
        ARG PROTOARCH=x86_64
    END
    RUN apt-get update && apt-get install -y unzip

    ARG PROTO_VERSION=3.19.4
    RUN curl -sSL -o protoc.zip https://github.com/protocolbuffers/protobuf/releases/download/v${PROTO_VERSION}/protoc-${PROTO_VERSION}-${TARGETOS}-${PROTOARCH}.zip
    RUN unzip protoc.zip -d /usr/local/

    ARG BUF_VERSION=1.2.1
    RUN curl -sSL "https://github.com/bufbuild/buf/releases/download/v${BUF_VERSION}/buf-$(uname -s)-$(uname -m)" -o "/usr/local/bin/buf"
        RUN chmod +x "/usr/local/bin/buf"

    RUN go install github.com/gogo/protobuf/protoc-gen-gogoslick@latest
    COPY tagging.proto buf.* .

gen-proto:
    FROM +proto-deps
    RUN buf generate
    SAVE ARTIFACT tagging.pb.go AS LOCAL tagging.pb.go

lint-proto:
    FROM +proto-deps
    RUN buf lint

breaking-proto:
    FROM +proto-deps
    RUN buf breaking --against "https://github.com/kentik/gohippo.git#branch=main"