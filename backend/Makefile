.PHONY: all api admin-api user-rpc paper-rpc rating-rpc news-rpc admin-rpc cron env-up env-down tidy start

COMPOSE_CMD := $(shell if docker compose version >/dev/null 2>&1; then echo "docker compose"; elif command -v docker-compose >/dev/null 2>&1; then echo "docker-compose"; else echo "docker-compose"; fi)

all: tidy api admin-api user-rpc paper-rpc rating-rpc news-rpc admin-rpc cron

# === Dependencies ===
tidy:
	go mod tidy

# === Infrastructure ===
env-up:
	$(COMPOSE_CMD) up -d

init-repl:
	bash deploy/mysql/init-replication.sh

env-down:
	$(COMPOSE_CMD) down

# === API Gateway ===
api:
	go run api/journal.go -f api/etc/journal-api.yaml

admin-api:
	go run admin-api/admin.go -f admin-api/etc/admin-api.yaml

# === RPC Services ===
user-rpc:
	go run rpc/user/user.go -f rpc/user/etc/user.yaml

paper-rpc:
	go run rpc/paper/paper.go -f rpc/paper/etc/paper.yaml

rating-rpc:
	go run rpc/rating/rating.go -f rpc/rating/etc/rating.yaml

news-rpc:
	go run rpc/news/news.go -f rpc/news/etc/news.yaml

admin-rpc:
	go run rpc/admin/admin.go -f rpc/admin/etc/admin.yaml

# === Unified Cron ===
cron:
	go run cmd/cron/main.go

# === Build ===
build-api:
	go build -o bin/api api/journal.go

build-admin-api:
	go build -o bin/admin-api admin-api/admin.go

build-user:
	go build -o bin/user-rpc rpc/user/user.go

build-paper:
	go build -o bin/paper-rpc rpc/paper/paper.go

build-rating:
	go build -o bin/rating-rpc rpc/rating/rating.go

build-news:
	go build -o bin/news-rpc rpc/news/news.go

build-admin-rpc:
	go build -o bin/admin-rpc rpc/admin/admin.go

build-cron:
	go build -o bin/cron cmd/cron/main.go

build-lifecycle:
	go build -o bin/lifecycle cmd/lifecycle/main.go

build: build-api build-admin-api build-user build-paper build-rating build-news build-admin-rpc build-cron build-lifecycle

# === Lifecycle Cron ===
lifecycle:
	go run cmd/lifecycle/main.go

# === Proto Generation ===
proto-user:
	export PATH=$$HOME/go/bin:$$PATH && goctl rpc protoc proto/user.proto --go_out=rpc/user --go-grpc_out=rpc/user --zrpc_out=rpc/user -m

proto-paper:
	export PATH=$$HOME/go/bin:$$PATH && goctl rpc protoc proto/paper.proto --go_out=rpc/paper --go-grpc_out=rpc/paper --zrpc_out=rpc/paper -m

proto-rating:
	export PATH=$$HOME/go/bin:$$PATH && goctl rpc protoc proto/rating.proto --go_out=rpc/rating --go-grpc_out=rpc/rating --zrpc_out=rpc/rating -m

proto-news:
	export PATH=$$HOME/go/bin:$$PATH && goctl rpc protoc proto/news.proto --go_out=rpc/news --go-grpc_out=rpc/news --zrpc_out=rpc/news -m

proto-all: proto-user proto-paper proto-rating proto-news

# === API Generation ===
api-gen:
	export PATH=$$HOME/go/bin:$$PATH && cd api && goctl api go -api journal.api -dir . -style goZero

start:
	bash start.sh dev
