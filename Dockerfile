FROM golang:1.8.3-alpine3.6
MAINTAINER Xue Bing <xuebing1110@gmail.com>

# repo
RUN cp /etc/apk/repositories /etc/apk/repositories.bak
RUN echo "http://mirrors.aliyun.com/alpine/v3.6/main/" > /etc/apk/repositories

# timezone
RUN apk update
RUN apk add --no-cache py-pip ansible openssh && pip install paramiko
RUN apk add --no-cache tzdata \
    && echo "Asia/Shanghai" > /etc/timezone \
    && ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime

# move to GOPATH
RUN mkdir -p /go/src/github.com/xuebing1110/hostadmin
COPY . $GOPATH/src/github.com/xuebing1110/hostadmin/
WORKDIR $GOPATH/src/github.com/xuebing1110/hostadmin

# copy config
COPY cluster-admin/etc/init.d/* /etc/init.d/
COPY cluster-admin/etc/playbook/ /app/
COPY cluster-admin/etc/sysconfig/* /etc/sysconfig/
COPY cluster-admin/etc/systemd/* /etc/systemd/system/

# build
RUN mkdir -p /app
RUN go build -o /app/cluster-admin cluster-admin/cmd/main.go

# example config
# COPY cluster-admin/cmd/conf.yaml /app/conf.yaml

EXPOSE 50051
WORKDIR /app
CMD ["/app/cluster-admin"]