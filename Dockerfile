FROM golang:latest as BUILDER

MAINTAINER TommyLike<tommylikehu@gmail.com>

# build binary
RUN mkdir -p /go/src/gitee.com/openeuler/hdc-task-manager
COPY . /go/src/gitee.com/openeuler/hdc-task-manager
RUN cd /go/src/gitee.com/openeuler/hdc-task-manager && CGO_ENABLED=1 go build -v -o ./hdc-task-manager main.go

# copy binary config and utils
FROM golang:latest
RUN mkdir -p /opt/app/ && mkdir -p /opt/app/conf/
COPY ./conf/product_app.conf /opt/app/conf/app.conf
# overwrite config yaml
COPY  --from=BUILDER /go/src/gitee.com/openeuler/hdc-task-manager/hdc-task-manager /opt/app

WORKDIR /opt/app/
ENTRYPOINT ["/opt/app/hdc-task-manager"]