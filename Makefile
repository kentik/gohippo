_GOPATH 			:= $(PWD)/../../../..
PROTOC := /usr/bin/protoc

export GOPATH := $(_GOPATH)
export PATH := $(_GOPATH)/bin:$(PATH)

.PHONY: all
all: codegen

$(_GOPATH)/src/github.com/gogo/protobuf/protoc-gen-gogoslick:
	go get github.com/gogo/protobuf/protoc-gen-gogoslick

$(_GOPATH)/bin/protoc-gen-gofast:
	go get github.com/gogo/protobuf/protoc-gen-gofast
	go install github.com/gogo/protobuf/protoc-gen-gofast

$(_GOPATH)/src/github.com/gogo/protobuf/proto:
	go get github.com/gogo/protobuf/proto

$(_GOPATH)/src/github.com/gogo/protobuf/jsonpb:
	go get github.com/gogo/protobuf/jsonpb

$(_GOPATH)/src/github.com/gogo/protobuf/protoc-gen-gogo:
	go get github.com/gogo/protobuf/protoc-gen-gogo

$(_GOPATH)/src/github.com/gogo/protobuf/gogoproto:
	go get github.com/gogo/protobuf/gogoproto

.PHONY: codegen
codegen: \
	$(_GOPATH)/src/github.com/gogo/protobuf/protoc-gen-gogoslick \
	$(_GOPATH)/bin/protoc-gen-gofast \
   	$(_GOPATH)/src/github.com/gogo/protobuf/proto \
	$(_GOPATH)/src/github.com/gogo/protobuf/jsonpb \
	$(_GOPATH)/src/github.com/gogo/protobuf/protoc-gen-gogo \
	$(_GOPATH)/src/github.com/gogo/protobuf/gogoproto

	$(PROTOC) -I=. \
	-I=$(_GOPATH)/src \
	-I=$(_GOPATH)/src/github.com/gogo/protobuf/protobuf \
	--gogoslick_out=.\
	Mgoogle/protobuf/any.proto=github.com/gogo/protobuf/types,\
	Mgoogle/protobuf/duration.proto=github.com/gogo/protobuf/types,\
	Mgoogle/protobuf/struct.proto=github.com/gogo/protobuf/types,\
	Mgoogle/protobuf/timestamp.proto=github.com/gogo/protobuf/types,\
	Mgoogle/protobuf/wrappers.proto=github.com/gogo/protobuf/types:. \
	  *.proto

.PHONY: test
test:
	go test -v ./...
