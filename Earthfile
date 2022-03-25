VERSION 0.6

FROM golang:1.17

WORKDIR /build

all:
    BUILD +build
    BUILD +test
    BUILD +gen-proto

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
    RUN apt-get update && apt-get install -y wget unzip

    RUN wget -O protoc.zip https://github.com/protocolbuffers/protobuf/releases/download/v3.13.0/protoc-3.13.0-${TARGETOS}-${PROTOARCH}.zip
    RUN unzip protoc.zip -d /usr/local/

    RUN VERSION="1.1.0" && \
            curl -sSL "https://github.com/bufbuild/buf/releases/download/v${VERSION}/buf-$(uname -s)-$(uname -m)" -o "/usr/local/bin/buf"
        RUN chmod +x "/usr/local/bin/buf"

    RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
    RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
    RUN go install github.com/gogo/protobuf/protoc-gen-gogoslick@latest
    COPY tagging.proto buf.* .

gen-proto:
    FROM +proto-deps
    RUN buf generate
    SAVE ARTIFACT tagging.pb.go AS LOCAL tagging.pb.go

lint-proto:
    FROM +proto-deps
    RUN buf lint