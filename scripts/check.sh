#!/usr/bin/env bash
# ---------------------------------------------------------------------------
# 一键体检：格式 / vet / 构建 / 测试
#   scripts/check.sh [子目录]     给子目录则只检查该目录所属 module
# ---------------------------------------------------------------------------
set -uo pipefail
SCRIPT_DIR="$(CDPATH= cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
MODULE_TOOL="$SCRIPT_DIR/go-modules.sh"
ROOT_DIR="$(CDPATH= cd -- "$SCRIPT_DIR/.." && pwd)"
p="${1:-}"
C_RESET=$'\033[0m'; C_CYAN=$'\033[36m'; C_GREEN=$'\033[32m'; C_RED=$'\033[31m'

fail=0
step() { printf '\n%s== %s ==%s\n' "$C_CYAN" "$1" "$C_RESET"; }

if [[ -n "$p" ]]; then
	modules=("$($MODULE_TOOL resolve "$p")")
else
	modules=($("$MODULE_TOOL" list))
fi

if [[ "${#modules[@]}" -eq 0 ]]; then
	printf '%s没有发现 Go module%s\n' "$C_RED" "$C_RESET"
	exit 2
fi

printf '\033[1mOriginium 体检%s  目标: %s\n' "$C_RESET" "${p:-全部 modules}"

for module in "${modules[@]}"; do
	if [[ "$module" == "." ]]; then
		module_dir="$ROOT_DIR"
	else
		module_dir="$ROOT_DIR/$module"
	fi
	printf '\n%s## module: %s ##%s\n' "$C_CYAN" "$module" "$C_RESET"

	step "gofmt"
	out="$(cd "$module_dir" && find . -type d \( -name .git -o -name vendor -o -name bin -o -name node_modules \) -prune -o -type f -name '*.go' -print0 | while IFS= read -r -d '' file; do gofmt -l "$file"; done)"
	if [[ -n "$out" ]]; then
		printf '%s✗ 以下文件未格式化:%s\n%s\n' "$C_RED" "$C_RESET" "$out"
		printf '  运行 make fmt p=%s 修复\n' "$module"
		fail=1
	else
		printf '%s✓ 格式 OK%s\n' "$C_GREEN" "$C_RESET"
	fi

	step "go vet"
	if (cd "$module_dir" && go vet ./...); then
		printf '%s✓ vet 通过%s\n' "$C_GREEN" "$C_RESET"
	else
		printf '%s✗ vet 发现问题%s\n' "$C_RED" "$C_RESET"
		fail=1
	fi

	step "go build"
	if (cd "$module_dir" && go build ./...); then
		printf '%s✓ 构建成功%s\n' "$C_GREEN" "$C_RESET"
	else
		printf '%s✗ 构建失败%s\n' "$C_RED" "$C_RESET"
		fail=1
	fi

	step "go test"
	if (cd "$module_dir" && go test ./...); then
		printf '%s✓ 测试通过%s\n' "$C_GREEN" "$C_RESET"
	else
		printf '%s✗ 测试失败%s\n' "$C_RED" "$C_RESET"
		fail=1
	fi
done

if [[ "$fail" -eq 0 ]]; then
	printf '\n%s🎉 全部通过%s\n\n' "$C_GREEN" "$C_RESET"
else
	printf '\n%s体检未通过，请修复上面标红的部分%s\n\n' "$C_RED" "$C_RESET"
fi
exit "$fail"