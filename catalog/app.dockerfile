FROM golang:1.26-alpine AS build
RUN apk --no-cache add ca-certificates
WORKDIR /go/src/github.com/akash1723tripathi/go-microservices
COPY go.mod go.sum ./
COPY catalog catalog
RUN go build -o /go/bin/app ./catalog/cmd/catalog

FROM alpine:3.22
WORKDIR /usr/bin
COPY --from=build /go/bin .
EXPOSE 8080
CMD ["app"]
