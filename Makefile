.PHONY: proto backend frontend docker-start docker-stop install run

# 基础设施
docker-start:
	cd docker && docker compose up -d

docker-stop:
	cd docker && docker compose down

# Proto 生成
proto:
	@if ! command -v protoc &> /dev/null; then \
		echo "protoc not found. Install: brew install protobuf"; \
	fi
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/auction.proto

# 后端
backend-deps:
	cd backend && go mod tidy

backend-run:
	cd backend && go run cmd/server/main.go

seed:
	cd backend && go run cmd/server/seed.go

# 前端
frontend-deps:
	cd frontend && npm install

frontend-run:
	cd frontend && npm run dev

frontend-build:
	cd frontend && npm run build

# 一键启动（开发）
dev: docker-start
	@echo "Waiting for PostgreSQL..."
	@sleep 3
	@echo "Starting backend..."
	@make backend-run &
	@echo "Starting frontend..."
	@make frontend-run

# 清理
clean:
	cd docker && docker compose down -v
	cd backend && rm -rf target/
	cd frontend && rm -rf dist/ node_modules/
