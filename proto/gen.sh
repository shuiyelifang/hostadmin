#!/bin/bash
echo "generate stub..."
protoc -I/usr/local/include \
    -I/Users/xuebing/WorkSpace/go/src \
    -I/Users/xuebing/WorkSpace/go/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
    -IHostManager/ \
    --go_out=plugins=grpc:hostManager \
    HostManager/hostManager.proto

## generate gateway
echo "generate gateway..."
protoc -I/usr/local/include \
    -I/Users/xuebing/WorkSpace/go/src \
    -I/Users/xuebing/WorkSpace/go/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
    -IHostManager/ \
    --grpc-gateway_out=logtostderr=true:hostManager \
    HostManager/hostManager.proto

## generate swagger
echo "generate swagger..."
protoc -I/usr/local/include \
    -I/Users/xuebing/WorkSpace/go/src \
    -I/Users/xuebing/WorkSpace/go/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
    -IHostManager/ \
    --swagger_out=logtostderr=true:hostManager \
    HostManager/hostManager.proto

echo "over!"