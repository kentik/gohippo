VERSION 0.6

FROM golang:1.17

WORKDIR /build

GO_RUN:
    COMMAND
    ARG --required cmd
    ARG GOCACHE=/go-cache
    RUN --mount=type=cache,target=$GOCACHE $cmd

deps:
    COPY go.mod go.sum .
    DO +GO_RUN --cmd="go mod download all"

test:
    FROM +deps
    COPY *.go .
    DO +GO_RUN --cmd="go test -v ./..."

codegen:
    FROM +proto-deps
    COPY tagging.proto .
    RUN protoc \
        --gogoslick_out=.\
        	Mgoogle/protobuf/any.proto=github.com/gogo/protobuf/types,\
        	Mgoogle/protobuf/duration.proto=github.com/gogo/protobuf/types,\
        	Mgoogle/protobuf/struct.proto=github.com/gogo/protobuf/types,\
        	Mgoogle/protobuf/timestamp.proto=github.com/gogo/protobuf/types,\
        	Mgoogle/protobuf/wrappers.proto=github.com/gogo/protobuf/types:. \
        	  *.proto

proto-deps:
    FROM golang:buster
    ARG TARGETOS
    ARG TARGETARCH
    RUN apt-get update && apt-get install -y wget unzip
    RUN echo ${TARGETARCH}
    RUN wget -O protoc.zip https://github.com/protocolbuffers/protobuf/releases/download/v3.13.0/protoc-3.13.0-${TARGETOS}-${TARGETARCH}.zip
    RUN unzip protoc.zip -d /usr/local/

    #RUN go install google.golang.org/protobuf/cmd/protoc-gen-go \
    #           google.golang.org/grpc/cmd/protoc-gen-go-grpc \
    #           github.com/gogo/protobuf/protoc-gen-gogoslick \
    #           github.com/gogo/protobuf/proto \
    #           github.com/gogo/protobuf/jsonpb

proto-deps-buf:
    FROM golang:buster
    ARG TARGETOS
    ARG TARGETARCH
    RUN VERSION="1.1.0" && \
            curl -sSL "https://github.com/bufbuild/buf/releases/download/v${VERSION}/buf-$(uname -s)-$(uname -m)" -o "/usr/local/bin/buf"
        RUN chmod +x "/usr/local/bin/buf"
    RUN pwd
    WORKDIR /work
    COPY go.mod go.sum .
    RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
    RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
    #RUN go install github.com/gogo/protobuf/protoc-gen-gogoslick@latest
    #RUN go install github.com/gogo/protobuf/protoc-gen-gogo@latest
    #RUN go install github.com/gogo/protobuf/protoc-gen-gofast@latest
    #RUN go get github.com/gogo/protobuf/proto@latest
    #RUN go get github.com/gogo/protobuf/jsonpb@latest
    RUN go get github.com/gogo/protobuf/gogoproto
    RUN ls -alh .

gen-buf:
    #FROM bufbuild/buf
    FROM +proto-deps-buf
    WORKDIR /go
    RUN ls -alhR .
    #ARG PATH="$PATH:$(go env GOPATH)/bin:/work"
    #RUN env
    COPY tagging.proto buf.gen.yaml .
    RUN buf generate

lint-proto:
    FROM bufbuild/buf
    RUN buf lint