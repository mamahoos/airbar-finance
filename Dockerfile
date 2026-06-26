# F0 skeleton — wire build steps when cmd/server exists.
FROM golang:1.22-alpine AS builder
WORKDIR /src
COPY go.mod ./
# COPY go.sum ./
# RUN go mod download
# COPY . .
# RUN CGO_ENABLED=0 go build -o /bin/airbar-finance ./cmd/server

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
# COPY --from=builder /bin/airbar-finance /usr/local/bin/airbar-finance
# EXPOSE 8080 50051
# ENTRYPOINT ["airbar-finance"]
