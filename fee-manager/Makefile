all:
	protoc -Iinterface/ -I$(GOPATH)/src/ -I/usr/local/include --grpc-gateway_out=logtostderr=true:interface/ --go_out=plugins=grpc:interface/ interface/fee.proto

proto:
	protoc -Iinterface/ -I$(GOPATH)/src/ -I$(GOPATH)/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis -I/usr/local/include --grpc-gateway_out=logtostderr=true:interface/ --go_out=plugins=grpc:interface/ interface/fee.proto
