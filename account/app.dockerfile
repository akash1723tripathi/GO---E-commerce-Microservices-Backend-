FROM golang:1.13-alpine3.11 AS build
RUN apk --no-cache add gcc g++ make ca-certificates
WORKDIR /go/src/github.com/akash1723tripathi/go-microservices
COPY go.mod go.sum ./
COPY account account
RUN GO111MODULE=on go build -o /go/bin/app ./account/cmd

FROM alpine:3.11
WORKDIR /usr/bin
COPY --from=build /go/bin .
EXPOSE 8080
CMD ["app"]
