protoc -I pkg/api/ pkg/api/*.proto --go_out=plugins=grpc:pkg/api
