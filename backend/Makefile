.PHONY: all api user-rpc paper-rpc rating-rpc news-rpc env-up env-down tidy

all: tidy api user-rpc paper-rpc rating-rpc news-rpc

# === Dependencies ===
tidy:
	go mod tidy

# === Infrastructure ===
env-up:
	docker-compose up -d

init-repl:
	bash deploy/mysql/init-replication.sh

env-down:
	docker-compose down

# === API Gateway ===
api:
	go run api/journal.go -f api/etc/journal-api.yaml

# === RPC Services ===
user-rpc:
	go run rpc/user/user.go -f rpc/user/etc/user.yaml

paper-rpc:
	go run rpc/paper/paper.go -f rpc/paper/etc/paper.yaml

rating-rpc:
	go run rpc/rating/rating.go -f rpc/rating/etc/rating.yaml

news-rpc:
	go run rpc/news/news.go -f rpc/news/etc/news.yaml

# === Build ===
build-api:
	go build -o bin/api api/journal.go

build-user:
	go build -o bin/user-rpc rpc/user/user.go

build-paper:
	go build -o bin/paper-rpc rpc/paper/paper.go

build-rating:
	go build -o bin/rating-rpc rpc/rating/rating.go

build-news:
	go build -o bin/news-rpc rpc/news/news.go

build-lifecycle:
	go build -o bin/lifecycle cmd/lifecycle/main.go

build: build-api build-user build-paper build-rating build-news build-lifecycle

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
