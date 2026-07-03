FROM golang:1.25-alpine AS builder

WORKDIR /build

COPY go.mod ./
COPY cmd/server ./cmd/server

RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /server ./cmd/server

FROM alpine:3.21

RUN apk add --no-cache ca-certificates

COPY --from=builder /server /server

EXPOSE 8080

ENTRYPOINT ["/server"]