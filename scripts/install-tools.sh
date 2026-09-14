#!/usr/bin/env bash
# ---------------------------------------------------------------------------
# 安装可选的增强工具（全部装到 $(go env GOPATH)/bin，不污染系统）
# 这些都是可选增强，不装也不影响 build / run / test。
# ---------------------------------------------------------------------------
set -euo pipefail

C_RESET=$'\033[0m'; C_CYAN=$'\033[36m'; C_GREEN=$'\033[32m'; C_DIM=$'\033[2m'

bin="$(go env GOPATH)/bin"
printf '%s安装到 %s%s\n\n' C_DIM "$bin" "$C_RESET"
mkdir -p "$bin"

install_one() {
	local name="$1" module="$2"; shift 2
	printf '%s▶ %s%s\n' "$C_CYAN" "$name" "$C_RESET"
	if command -v "$name" >/dev/null 2>&1; then
		printf '  已存在，跳过\n\n'
		return
	fi
	if go install "$@" "$module"; then
		printf '%s  ✓ 完成%s\n\n' "$C_GREEN" "$C_RESET"
	else
		printf '  ✗ 失败（可能网络受限），不影响其他功能\n\n'
	fi
}

# 静态检查（make lint 会用它，能揪出 nil map、漏关 resp.Body 之类经典坑）
install_one golangci-lint github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest

# 热重载（make watch 检测到它就用，比内置轮询体验好）
install_one air github.com/air-verse/air@latest

# 调试器（VS Code launch.json 里的断点调试需要它）
install_one dlv github.com/go-delve/delve/cmd/dlv@latest

# 更好看的测试输出
install_one gotestsum gotest.tools/gotestsum@latest

case ":$PATH:" in
	*":$bin:"*) ;;
	*)
		printf '\033[33m提示: %s 不在 PATH 里，加到 shell 配置：%s\n' "$bin" "$C_RESET"
		printf '  export PATH="$PATH:%s"\n\n' "$bin"
		;;
esac

printf '装完了。验证一下: make info\n\n'