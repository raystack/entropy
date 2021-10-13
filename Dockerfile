FROM golang:1.16-alpine3.13 AS builder
WORKDIR /go/src/github.com/odpf/entropy
COPY . .
RUN apk add make bash git
RUN make build

FROM alpine:3.13
RUN apk --no-cache add ca-certificates bash
WORKDIR /root/
EXPOSE 8080
COPY --from=builder /go/src/github.com/odpf/entropy/build/entropy .
ENTRYPOINT ["./entropy"]
