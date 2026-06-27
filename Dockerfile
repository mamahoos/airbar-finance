FROM golang:1.26-alpine AS builder
WORKDIR /src
RUN apk add --no-cache git ca-certificates
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -buildvcs=false -o /bin/airbar-finance ./cmd/server

FROM alpine:3.24
RUN apk add --no-cache ca-certificates
COPY --from=builder /bin/airbar-finance /usr/local/bin/airbar-finance
EXPOSE 8080 50051
ENTRYPOINT ["airbar-finance"]
