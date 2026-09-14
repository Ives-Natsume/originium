#!/usr/bin/env bash
# ---------------------------------------------------------------------------
# 一键体检：格式 / vet / 构建 / 测试
#   scripts/check.sh [子目录]     给子目录则只检查该目录
# ---------------------------------------------------------------------------
set -uo pipefail

p="${1:-}"
pkg="./..."
[[ -n "$p" ]] && pkg="./${p#./}/..."

C_RESET=$'\033[0m'; C_CYAN=$'\033[36m'; C_GREEN=$'\033[32m'; C_RED=$'\033[31m'

fail=0
step() { printf '\n%s== %s ==%s\n' "$C_CYAN" "$1" "$C_RESET"; }

printf '\033[1mOriginium 体检%s  目标: %s\n' "$C_RESET" "$pkg"

# ---- 1. gofmt ------------------------------------------------------------
step "gofmt"
if [[ -n "$p" ]]; then
	out="$(gofmt -l "$p")"
else
	out="$(gofmt -l .)"
fi
if [[ -n "$out" ]]; then
	printf '%s✗ 以下文件未格式化:%s\n%s\n' "$C_RED" "$C_RESET" "$out"
	printf '  运行 make fmt 修复\n'
	fail=1
else
	printf '%s✓ 格式 OK%s\n' "$C_GREEN" "$C_RESET"
fi

# ---- 2. go vet -----------------------------------------------------------
step "go vet"
if go vet "$pkg"; then
	printf '%s✓ vet 通过%s\n' "$C_GREEN" "$C_RESET"
else
	printf '%s✗ vet 发现问题%s\n' "$C_RED" "$C_RESET"
	fail=1
fi

# ---- 3. build ------------------------------------------------------------
step "go build"
if go build "$pkg"; then
	printf '%s✓ 构建成功%s\n' "$C_GREEN" "$C_RESET"
else
	printf '%s✗ 构建失败%s\n' "$C_RED" "$C_RESET"
	fail=1
fi

# ---- 4. test -------------------------------------------------------------
step "go test"
if go test "$pkg"; then
	printf '%s✓ 测试通过%s\n' "$C_GREEN" "$C_RESET"
else
	printf '%s✗ 测试失败%s\n' "$C_RED" "$C_RESET"
	fail=1
fi

# ---- 汇总 ----------------------------------------------------------------
if [[ "$fail" -eq 0 ]]; then
	printf '\n%s🎉 全部通过%s\n\n' "$C_GREEN" "$C_RESET"
else
	printf '\n%s体检未通过，请修复上面标红的部分%s\n\n' "$C_RED" "$C_RESET"
fi
exit "$fail"