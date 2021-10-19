FROM golang:1.17.2 AS builder
WORKDIR $GOPATH/say-hello-server
COPY . . 
RUN go mod download
# cross compilation enabled (by default it is built for native system only)
RUN CGO_ENABLED=0 go build -o hello . 

FROM alpine:3.14
COPY --from=builder /go/say-hello-server/hello /
CMD [ "./hello" ]