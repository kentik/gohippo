PROTOC := protoc

.PHONY: all
all: codegen

$(GOPATH)/src/github.com/gogo/protobuf/protoc-gen-gogoslick:
	go get github.com/gogo/protobuf/protoc-gen-gogoslick

$(GOPATH)/bin/protoc-gen-gofast:
	go get github.com/gogo/protobuf/protoc-gen-gofast
	go install github.com/gogo/protobuf/protoc-gen-gofast

$(GOPATH)/src/github.com/gogo/protobuf/proto:
	go get github.com/gogo/protobuf/proto

$(GOPATH)/src/github.com/gogo/protobuf/jsonpb:
	go get github.com/gogo/protobuf/jsonpb

$(GOPATH)/src/github.com/gogo/protobuf/protoc-gen-gogo:
	go get github.com/gogo/protobuf/protoc-gen-gogo

$(GOPATH)/src/github.com/gogo/protobuf/gogoproto:
	go get github.com/gogo/protobuf/gogoproto

.PHONY: codegen
codegen: \
	$(GOPATH)/src/github.com/gogo/protobuf/protoc-gen-gogoslick \
	$(GOPATH)/bin/protoc-gen-gofast \
   	$(GOPATH)/src/github.com/gogo/protobuf/proto \
	$(GOPATH)/src/github.com/gogo/protobuf/jsonpb \
	$(GOPATH)/src/github.com/gogo/protobuf/protoc-gen-gogo \
	$(GOPATH)/src/github.com/gogo/protobuf/gogoproto

	$(PROTOC) -I=. \
	-I=$(GOPATH)/src \
	-I=$(GOPATH)/src/github.com/gogo/protobuf/protobuf \
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

.PHONY: update
update:
	go get -u all
	go mod tidy
