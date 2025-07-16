FROM golang:alpine3.20 AS builder

RUN apk add --no-cache --virtual .build-deps gcc musl-dev openssl git

ENV GO111MODULE=on
RUN mkdir /go/src/github.com
RUN mkdir /go/src/github.com/cheetahfox
COPY ./ /go/src/github.com/cheetahfox/longping

WORKDIR /go/src/github.com/cheetahfox/longping

RUN go build

FROM alpine:3.22.1

RUN apk add --no-cache ca-certificates 

COPY --from=builder /go/src/github.com/cheetahfox/longping/longping /longping

RUN chmod +x /longping
CMD /longping
EXPOSE 3000



