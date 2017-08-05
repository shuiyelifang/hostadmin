FROM golang:1.8.3-onbuild
MAINTAINER Xue Bing <xuebing1110@gmail.com>

# timezone
RUN cp -f /usr/share/zoneinfo/Asia/Shanghai /etc/localtime

# move to GOPATH
RUN mkdir -p /go/src/github.com/xuebing1110/hostadmin
COPY . $GOPATH/src/github.com/xuebing1110/hostadmin/
WORKDIR $GOPATH/src/github.com/xuebing1110/hostadmin

# build
RUN mkdir -p /app
RUN go build -o /app/cluster-admin cluster-admin/cmd/main.go

# example config
# COPY cluster-admin/cmd/conf.yaml /app/conf.yaml

EXPOSE 50051
WORKDIR /app
CMD ["/app/cluster-admin"]