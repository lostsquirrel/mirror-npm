FROM registry.lisong.pub:28500/sunrise/golang:1.16-buster AS builder

ENV GO111MODULE=on
ENV GOPROXY=https://goproxy.cn,direct
ENV CGO_ENABLED=0
ADD . /dist
WORKDIR /dist
RUN go get -v all
RUN go build \
        -a -installsuffix cgo \
        -o bootstrap main.go

FROM scratch

COPY --from=builder /dist/bootstrap /
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
ENV TZ=Asia/Shanghai
ENV LANG=C.UTF-8
ENTRYPOINT ["/bootstrap"]
