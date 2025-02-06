FROM golang:alpine3.20 AS builder

RUN apk add --no-cache --virtual .build-deps gcc musl-dev openssl git

ENV GO111MODULE=on
RUN mkdir /go/src/github.com
RUN mkdir /go/src/github.com/cheetahfox
COPY ./ /go/src/github.com/cheetahfox/embeded-ping

WORKDIR /go/src/github.com/cheetahfox/embeded-ping

RUN go build

FROM alpine:3.20.2

RUN apk add --no-cache ca-certificates 

COPY --from=builder /go/src/github.com/cheetahfox/embeded-ping/embeded-ping /embeded-ping

RUN chmod +x /embeded-ping
CMD /embeded-ping
EXPOSE 3000



