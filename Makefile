SHELL := /bin/sh
GO ?= go
BUN ?= ~/.bun/bin/bun
MAKEFLAGS += --no-print-directory

.DEFAULT_GOAL := help

.PHONY: help deps deps-ci dev build test check \
	backend-run backend-test backend-vet backend-fix backend-backfill-boss-kills backend-check-boss-kills \
	frontend-dev frontend-build frontend-preview frontend-test hooks-install

help: ## 显示可用命令
	@awk 'BEGIN {FS = ":.*## "}; /^[a-zA-Z0-9_.-]+:.*## / {printf "%-18s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

deps: ## 安装前端依赖
	$(BUN) --cwd=frontend install

deps-ci: ## 以 CI 模式安装前端依赖
	$(BUN) --cwd=frontend ci

dev: ## 同时启动后端和前端开发环境
	@backend_pid=''; frontend_pid=''; status=0; \
	trap 'test -n "$$backend_pid" && kill $$backend_pid 2>/dev/null || true; test -n "$$frontend_pid" && kill $$frontend_pid 2>/dev/null || true' INT TERM EXIT; \
	$(GO) -C backend run ./cmd/server & backend_pid=$$!; \
	$(BUN) --cwd=frontend run dev & frontend_pid=$$!; \
	while :; do \
		if ! kill -0 $$backend_pid 2>/dev/null; then \
			wait $$backend_pid || status=$$?; \
			break; \
		fi; \
		if ! kill -0 $$frontend_pid 2>/dev/null; then \
			wait $$frontend_pid || status=$$?; \
			break; \
		fi; \
		sleep 1; \
	done; \
	kill $$backend_pid $$frontend_pid 2>/dev/null || true; \
	wait $$backend_pid 2>/dev/null || true; \
	wait $$frontend_pid 2>/dev/null || true; \
	exit $$status

build: frontend-build ## 构建前端产物到 backend/public

test: backend-test ## 运行后端测试

check: test backend-vet frontend-test build ## 执行 CI 校验

backend-run: ## 单独启动后端服务
	$(GO) -C backend run ./cmd/server

backend-test: ## 运行后端测试
	$(GO) -C backend test ./...

backend-vet: ## 运行 go vet
	$(GO) -C backend vet ./...

backend-fix: ## 运行 go fix
	$(GO) -C backend fix ./...

backend-backfill-boss-kills: ## 回填 Boss 击杀统计到 Redis
	$(GO) -C backend run ./cmd/backfillbosskills

backend-check-boss-kills: ## 检查 Redis 中的 Boss 击杀统计
	$(GO) -C backend run ./cmd/checkbosskills

frontend-dev: ## 单独启动前端开发服务器
	$(BUN) --cwd=frontend run dev

frontend-build: ## 构建前端产物
	$(BUN) --cwd=frontend run build

frontend-test: ## 运行前端测试
	$(BUN) --cwd=frontend run test

frontend-preview: ## 预览前端产物
	$(BUN) --cwd=frontend run preview

hooks-install: ## 安装 Git hooks（需要本地已安装 lefthook）
	lefthook install
