## generate stub
```sh
protoc -I/usr/local/include \
    -I/Users/xuebing/WorkSpace/go/src \
    -I/Users/xuebing/WorkSpace/go/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
    --go_out=plugins=grpc:hostManager
    -IHostManager/ HostManager/hostManager.proto
```

## generate gateway
```sh
protoc -I/usr/local/include \
    -I/Users/xuebing/WorkSpace/go/src \
    -I/Users/xuebing/WorkSpace/go/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
    --grpc-gateway_out=logtostderr=true:hostManager \
    -IHostManager/ HostManager/hostManager.proto

```

## generate swagger
```sh
protoc -I/usr/local/include \
    -I/Users/xuebing/WorkSpace/go/src \
    -I/Users/xuebing/WorkSpace/go/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
    --swagger_out=logtostderr=true:hostManager \
    -IHostManager/ HostManager/hostManager.proto
```