#!/usr/bin/env bash
# ---------------------------------------------------------------------------
# 显示环境 + 课程结构 + 包/测试统计
# ---------------------------------------------------------------------------
set -uo pipefail

C_RESET=$'\033[0m'; C_CYAN=$'\033[36m'; C_BOLD=$'\033[1m'; C_DIM=$'\033[2m'
SCRIPT_DIR="$(CDPATH= cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
MODULE_TOOL="$SCRIPT_DIR/go-modules.sh"
ROOT_DIR="$(CDPATH= cd -- "$SCRIPT_DIR/.." && pwd)"

printf '\n%sOriginium · 环境信息%s\n' "$C_BOLD" "$C_RESET"
printf '%s模块%s      %s\n' "$C_DIM" "$C_RESET" "$(head -1 "$ROOT_DIR/go.mod" | awk '{print $2}')"
printf '%sGo 版本%s   %s\n' "$C_DIM" "$C_RESET" "$(go version)"
printf '%sGOPATH%s    %s\n' "$C_DIM" "$C_RESET" "$(go env GOPATH)"
printf '%s代理%s      %s\n' "$C_DIM" "$C_RESET" "$(go env GOPROXY)"

printf '\n%s课程目录%s\n' "$C_BOLD" "$C_RESET"
for d in [0-9][0-9]-*/; do
	[[ -d "$d" ]] || continue
	n=$(find "$d" -name '*.go' | wc -l)
	readme="—"
	[[ -f "$d/README.md" ]] && readme="README.md"
	printf '  %s%-22s%s %3s 个 .go 文件   %s\n' "$C_CYAN" "${d%/}" "$C_RESET" "$n" "$readme"
done

printf '\n%s包统计%s\n' "$C_BOLD" "$C_RESET"
module_count=0
pkgs=0
for module in $($MODULE_TOOL list); do
	if [[ "$module" == "." ]]; then module_dir="$ROOT_DIR"; else module_dir="$ROOT_DIR/$module"; fi
	module_pkgs=$(cd "$module_dir" && go list ./... 2>/dev/null | wc -l)
	module_name=$(cd "$module_dir" && go list -m -f '{{.Path}}' 2>/dev/null || printf '未知')
	printf '  %-22s %3s 个包  %s\n' "$module" "$module_pkgs" "$module_name"
	pkgs=$((pkgs + module_pkgs))
	module_count=$((module_count + 1))
done
printf '  包总数      %s\n' "$pkgs"
printf '  module 数量  %s\n' "$module_count"

printf '\n%s测试用例%s\n' "$C_BOLD" "$C_RESET"
tests=0
for module in $($MODULE_TOOL list); do
	if [[ "$module" == "." ]]; then module_dir="$ROOT_DIR"; else module_dir="$ROOT_DIR/$module"; fi
	module_tests=$(cd "$module_dir" && go test -list '.*' ./... 2>/dev/null | grep -cE '^(Test|Example|Benchmark)' || true)
	tests=$((tests + module_tests))
done
printf '  数量        %s\n' "$tests"
printf '  跑一遍      make test\n'

printf '\n%s可选增强工具%s\n' "$C_BOLD" "$C_RESET"
for t in golangci-lint air dlv gotestsum; do
	if command -v "$t" >/dev/null 2>&1; then
		printf '  %s✓%s %s\n' "$C_CYAN" "$C_RESET" "$t"
	else
		printf '  %s·%s %-14s 未安装（make tools）\n' "$C_DIM" "$C_RESET" "$t"
	fi
done
printf '\n'