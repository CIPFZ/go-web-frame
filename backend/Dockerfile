FROM golang:1.25.2-alpine AS builder

WORKDIR /app
ENV GOPROXY=https://goproxy.cn,direct

# Install deps for sqlite/cgo if needed, though mostly using pure go
RUN apk add --no-cache gcc musl-dev

COPY . .

RUN go mod download
RUN go build -o main ./cmd/server
RUN go build -o migrate ./cmd/migrate/main.go

FROM alpine:latest
WORKDIR /app

# Required for timezones and ca-certs
RUN apk --no-cache add tzdata ca-certificates

COPY --from=builder /app/main .
COPY --from=builder /app/migrate .
COPY --from=builder /app/configs ./configs
COPY start.sh .

RUN sed -i 's/\r$//' start.sh && chmod +x start.sh

EXPOSE 8080

CMD ["./start.sh"]
