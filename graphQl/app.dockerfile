FROM golang:1.26-alpine AS build
RUN apk --no-cache add ca-certificates
WORKDIR /src
COPY go.mod go.sum ./
COPY account account
COPY catalog catalog
COPY order order
COPY graphQl graphQl
RUN go build -o /go/bin/app ./graphQl/cmd

FROM alpine:3.22
COPY --from=build /go/bin/app /usr/bin/app
EXPOSE 8080
CMD ["/usr/bin/app"]
