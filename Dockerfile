FROM registry.lisong.pub:28500/hub/library/golang:1.24 AS builder

ENV GOPROXY=https://goproxy.cn,direct

ADD . /dist
WORKDIR /dist
RUN go get -v all
RUN go build \
        -o bootstrap main.go

FROM scratch

COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /dist/bootstrap /
ENV TZ=Asia/Shanghai
ENV LANG=C.UTF-8
ENTRYPOINT ["/bootstrap"]
