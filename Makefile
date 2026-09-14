# ============================================================================
#  Originium · Go 学习工作台
#
#  常用姿势：
#    make help                          看所有命令
#    make test                          跑全仓测试
#    make run p=01-basics/hello         运行某个 main 包
#    make run p=01-basics/hello a=Neo   运行并传参（第 3 个参数直接写）
#    make watch p=03-http/01-server     改代码自动重编译/重启
#    make new p=01-basics/newthing      新建练习骨架
#    make lib p=01-basics/mypkg         新建库包（带 _test.go 表驱动模板）
# ============================================================================

MODULE  := $(shell head -1 go.mod | awk '{print $$2}')
BIN_DIR := bin

# p= 指定子目录；不传则作用于全仓
PKG := $(if $(p),./$(p)/...,./...)

# 允许 `make run p=xxx myarg` 这种写法：第 2 个 goal 视为程序参数
a     := $(word 2,$(MAKECMDGOALS))
ARGS  := $(a)

.DEFAULT_GOAL := help
SHELL := /bin/bash

# ---------------------------------------------------------------------------
#  帮助
# ---------------------------------------------------------------------------
.PHONY: help
help: ## 显示所有可用命令
	@printf "\n\033[1mOriginium · Go 学习工作台\033[0m  (%s)\n\n" "$(MODULE)"
	@printf "用法: \033[36mmake <命令> [p=<子目录>] [a=<程序参数>]\033[0m\n"
	@printf "      p 缺省时命令作用于整个仓库\n\n"
	@awk 'BEGIN{FS=":.*##"} /^[a-zA-Z0-9_.-]+:.*##/ {printf "  \033[36m%-14s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
	@printf "\n示例: make run p=01-basics/hello a=Neo\n\n"

# ---------------------------------------------------------------------------
#  构建 / 运行
# ---------------------------------------------------------------------------
.PHONY: build
build: ## 编译所有包（只检查，不产出二进制）
	go build ./...

.PHONY: run
run: ## 运行某个 main 包: make run p=01-basics/hello [a=参数]
	@if [ -z "$(p)" ]; then printf "请指定包路径: make run p=01-basics/hello\n"; exit 2; fi
	go run ./$(p) $(ARGS)

.PHONY: build-bin
build-bin: ## 编译成二进制放到 bin/: make build-bin p=06-rpg-server/cmd/server
	@if [ -z "$(p)" ]; then printf "请指定包路径: make build-bin p=06-rpg-server/cmd/server\n"; exit 2; fi
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(notdir $(p)) ./$(p)
	@printf "已生成 \033[36m%s\033[0m\n" "$(BIN_DIR)/$(notdir $(p))"

.PHONY: watch
watch: ## 热重载：改代码自动重编译+重启（main 包）/ 重跑测试（其他）
	@./scripts/watch.sh "$(p)"

# ---------------------------------------------------------------------------
#  测试
# ---------------------------------------------------------------------------
.PHONY: test
test: ## 跑测试
	go test $(PKG)

.PHONY: test-v
test-v: ## 跑测试（详细输出）
	go test -v $(PKG)

.PHONY: test-race
test-race: ## 跑测试并开启竞态检测（并发章节必备）
	go test -race $(PKG)

.PHONY: test-cover
test-cover: ## 跑测试并输出覆盖率
	go test -cover $(PKG)

.PHONY: cover-html
cover-html: ## 生成覆盖率可视化报告 coverage.html
	@go test -coverprofile=coverage.out $(PKG)
	@go tool cover -html=coverage.out -o coverage.html
	@printf "已生成 \033[36mcoverage.html\033[0m，用编辑器打开预览\n"

.PHONY: bench
bench: ## 跑基准测试
	go test -bench=. -benchmem -run=^$$ $(PKG)

.PHONY: test-list
test-list: ## 列出仓库里所有测试用例
	@go test -list '.*' $(PKG) 2>/dev/null | grep -E '^(Test|Example|Benchmark)' | sort -u || true

# ---------------------------------------------------------------------------
#  代码质量
# ---------------------------------------------------------------------------
.PHONY: fmt
fmt: ## 自动格式化 + 整理依赖
	gofmt -w .
	go mod tidy

.PHONY: vet
vet: ## go vet 静态检查
	go vet $(PKG)

.PHONY: lint
lint: ## gofmt 检查 + go vet（可用 make tools 装 golangci-lint 增强）
	@out="$$(gofmt -l .)"; \
	if [ -n "$$out" ]; then printf "\033[31m以下文件未格式化（跑 make fmt 修复）:\033[0m\n%s\n" "$$out"; exit 1; fi
	@printf "\033[32m✓ gofmt OK\033[0m\n"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		printf "\033[33m提示: 未安装 golangci-lint，仅执行 go vet。\033[0m\n"; \
		go vet $(PKG); \
	fi

.PHONY: check
check: ## 一键体检：格式 + vet + build + test
	@./scripts/check.sh "$(p)"

# ---------------------------------------------------------------------------
#  脚手架
# ---------------------------------------------------------------------------
.PHONY: new
new: ## 新建可运行练习骨架: make new p=01-basics/mytopic
	@if [ -z "$(p)" ]; then printf "请指定目录: make new p=01-basics/mytopic\n"; exit 2; fi
	@./scripts/new.sh "$(p)" app

.PHONY: lib
lib: ## 新建库包骨架（含表驱动测试模板）: make lib p=01-basics/mypkg
	@if [ -z "$(p)" ]; then printf "请指定目录: make lib p=01-basics/mypkg\n"; exit 2; fi
	@./scripts/new.sh "$(p)" lib

# ---------------------------------------------------------------------------
#  依赖 / 环境
# ---------------------------------------------------------------------------
.PHONY: deps tidy
deps tidy: ## go mod tidy
	go mod tidy

.PHONY: tools
tools: ## 安装可选的增强工具（golangci-lint / air / dlv）
	@./scripts/install-tools.sh

.PHONY: docker-up
docker-up: ## 启动 MySQL + Redis（需要 docker）
	docker compose -f deploy/docker-compose.yml up -d

.PHONY: docker-down
docker-down: ## 停止 MySQL + Redis
	docker compose -f deploy/docker-compose.yml down

.PHONY: info
info: ## 显示 Go 环境、目录结构、包与测试统计
	@./scripts/info.sh

.PHONY: clean
clean: ## 清理构建产物与测试缓存
	go clean -cache -testcache
	rm -rf $(BIN_DIR) coverage.out coverage.html
	@printf "已清理\n"

# 允许 `make run p=x 参数` 中的 "参数" 不被当成目标
ifneq ($(a),)
$(a):
	@:
endif