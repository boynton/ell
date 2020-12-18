FROM golang:alpine

ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=arm64

WORKDIR /go/src/ell
COPY ./lib .

RUN go get github.com/boynton/ell/...

CMD ["time", "ell", "bench"]
