FROM golang:1.8.3-alpine3.6
MAINTAINER Xue Bing <xuebing1110@gmail.com>


# repo
RUN cp /etc/apk/repositories /etc/apk/repositories.bak
RUN echo "http://mirrors.aliyun.com/alpine/v3.6/main/" > /etc/apk/repositories

# timezone
RUN apk update && apk add --no-cache ansible \
    &&  apk add --no-cache tzdata \
    && cp -r -f /usr/share/zoneinfo/Hongkong /etc/localtime \\
    && echo -ne "Alpine Linux 3.6 image. (`uname -rsv`)\n" >> /root/.built

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