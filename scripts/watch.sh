#!/usr/bin/env bash
# ---------------------------------------------------------------------------
# 热重载（无外部依赖）
#   scripts/watch.sh <目录>
#
#   * 目录是 main 包  -> 改代码后自动重新编译并重启进程
#   * 其他包          -> 改代码后自动重跑 go test
#
#   如果安装了 air，main 包会优先交给 air 处理（体验更好，见 make tools）。
#   Ctrl-C 退出。
# ---------------------------------------------------------------------------
set -uo pipefail

dir="${1:-}"
if [[ -z "$dir" ]]; then
	printf '用法: scripts/watch.sh <目录>   例: scripts/watch.sh 03-http/01-server\n' >&2
	exit 2
fi
if [[ ! -d "$dir" ]]; then
	printf '\033[31m目录不存在: %s\033[0m\n' "$dir" >&2
	exit 2
fi

# 转成 go 可用的相对路径
target="./${dir#./}"

C_RESET=$'\033[0m'; C_DIM=$'\033[2m'; C_GREEN=$'\033[32m'
C_RED=$'\033[31m'; C_CYAN=$'\033[36m'; C_YELLOW=$'\033[33m'

# ---- 判断包类型 ----------------------------------------------------------
is_main=false
if go list -f '{{.Name}}' "$target" >/dev/null 2>&1; then
	[[ "$(go list -f '{{.Name}}' "$target" 2>/dev/null)" == "main" ]] && is_main=true
else
	printf '\033[31m无法解析包: %s（先确认目录里有 .go 文件）\033[0m\n' "$target" >&2
	exit 2
fi

# ---- air 优先 ------------------------------------------------------------
if $is_main && command -v air >/dev/null 2>&1; then
	printf '%s▶ 检测到 air，使用 air 热重载 %s%s\n' "$C_CYAN" "$target" "$C_RESET"
	exec air --build.cmd "go build -o ./bin/.watch-air $target" --build.bin "./bin/.watch-air"
fi

# ---- 快照函数 ------------------------------------------------------------
snapshot() {
	find "$dir" -type f -name '*.go' -printf '%T@ %p\n' 2>/dev/null | sort | md5sum
}

tmpdir="$(mktemp -d)"
pid=""

cleanup() {
	[[ -n "$pid" ]] && kill "$pid" 2>/dev/null
	rm -rf "$tmpdir"
}
trap 'cleanup; printf "\n%s已退出监视%s\n" "$C_DIM" "$C_RESET"; exit 0' INT TERM

if $is_main; then
	bin="$tmpdir/app"
	build_and_run() {
		[[ -n "$pid" ]] && kill "$pid" 2>/dev/null; pid=""
		if ! go build -o "$bin" "$target" 2>&1; then
			printf '%s✗ 编译失败，等待修改…%s\n' "$C_RED" "$C_RESET"
			return
		fi
		"$bin" &
		pid=$!
		printf '%s✓ 已启动 (pid %s)%s\n' "$C_GREEN" "$pid" "$C_RESET"
	}
	printf '%s▶ 热重载 main 包: %s（Ctrl-C 退出）%s\n' "$C_CYAN" "$target" "$C_RESET"
else
	run_tests() {
		printf '%s▶ go test %s%s\n' "$C_CYAN" "$target" "$C_RESET"
		go test "$target" || printf '%s✗ 测试未通过%s\n' "$C_YELLOW" "$C_RESET"
	}
	printf '%s▶ 目录不是 main 包，改为监视并重跑测试: %s%s\n' "$C_CYAN" "$target" "$C_RESET"
fi

prev="$(snapshot)"
if $is_main; then build_and_run; else run_tests; fi

while true; do
	sleep 1
	cur="$(snapshot)"
	[[ "$cur" == "$prev" ]] && continue
	prev="$cur"
	printf '\n%s──── 检测到改动 %s ────%s\n' "$C_DIM" "$(date +%H:%M:%S)" "$C_RESET"
	if $is_main; then build_and_run; else run_tests; fi
done