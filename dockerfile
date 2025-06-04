FROM golang:latest AS builder

WORKDIR /app
COPY . .
RUN go mod tidy
RUN CGO_ENABLED=0 go build -o seckill-service .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /app/seckill-service .
# 如果有配置文件，记得也复制它
COPY configs.json .

EXPOSE 8080
CMD ["./seckill-service"]