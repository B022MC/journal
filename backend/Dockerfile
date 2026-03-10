# ============================================================
# Multi-stage Dockerfile for Journal Backend Services
# Usage:
#   docker build --target api        -t journal-api .
#   docker build --target user-rpc   -t journal-user-rpc .
#   docker build --target paper-rpc  -t journal-paper-rpc .
#   docker build --target rating-rpc -t journal-rating-rpc .
#   docker build --target news-rpc   -t journal-news-rpc .
#   docker build --target admin-api  -t journal-admin-api .
#   docker build --target admin-rpc  -t journal-admin-rpc .
#   docker build --target lifecycle  -t journal-lifecycle .
# ============================================================

# ---- Build Stage ----
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /src
COPY go.mod go.sum ./
ENV GOPROXY=https://goproxy.cn,direct
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/api          api/journal.go
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/user-rpc     rpc/user/user.go
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/paper-rpc    rpc/paper/paper.go
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/rating-rpc   rpc/rating/rating.go
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/news-rpc     rpc/news/news.go
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/admin-api    admin-api/admin.go
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/admin-rpc    rpc/admin/admin.go
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/lifecycle    cmd/lifecycle/main.go

# ---- Runtime: API Gateway ----
FROM alpine:3.20 AS api
RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /out/api /app/api
COPY api/etc /app/etc
WORKDIR /app
EXPOSE 8888
CMD ["./api", "-f", "etc/journal-api.yaml"]

# ---- Runtime: User RPC ----
FROM alpine:3.20 AS user-rpc
RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /out/user-rpc /app/user-rpc
COPY rpc/user/etc /app/etc
WORKDIR /app
EXPOSE 9001
CMD ["./user-rpc", "-f", "etc/user.yaml"]

# ---- Runtime: Paper RPC ----
FROM alpine:3.20 AS paper-rpc
RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /out/paper-rpc /app/paper-rpc
COPY rpc/paper/etc /app/etc
WORKDIR /app
EXPOSE 9002
CMD ["./paper-rpc", "-f", "etc/paper.yaml"]

# ---- Runtime: Rating RPC ----
FROM alpine:3.20 AS rating-rpc
RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /out/rating-rpc /app/rating-rpc
COPY rpc/rating/etc /app/etc
WORKDIR /app
EXPOSE 9003
CMD ["./rating-rpc", "-f", "etc/rating.yaml"]

# ---- Runtime: News RPC ----
FROM alpine:3.20 AS news-rpc
RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /out/news-rpc /app/news-rpc
COPY rpc/news/etc /app/etc
WORKDIR /app
EXPOSE 9004
CMD ["./news-rpc", "-f", "etc/news.yaml"]

# ---- Runtime: Lifecycle Cron ----
FROM alpine:3.20 AS lifecycle
RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /out/lifecycle /app/lifecycle
WORKDIR /app
CMD ["./lifecycle"]

# ---- Runtime: Admin API Gateway ----
FROM alpine:3.20 AS admin-api
RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /out/admin-api /app/admin-api
COPY admin-api/etc /app/etc
WORKDIR /app
EXPOSE 8889
CMD ["./admin-api", "-f", "etc/admin-api.yaml"]

# ---- Runtime: Admin RPC ----
FROM alpine:3.20 AS admin-rpc
RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /out/admin-rpc /app/admin-rpc
COPY rpc/admin/etc /app/etc
WORKDIR /app
EXPOSE 9005
CMD ["./admin-rpc", "-f", "etc/admin.yaml"]
