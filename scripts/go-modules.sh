#!/usr/bin/env bash
# ---------------------------------------------------------------------------
# 仓库内 Go module 工具。
#
# 根模块里的课程练习和 projects/ 下的独立项目可以共存。所有命令都先
# 找到目标所属的最近 go.mod，再使用 go -C 在对应模块内执行。
# ---------------------------------------------------------------------------
set -uo pipefail

SCRIPT_DIR="$(CDPATH= cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(CDPATH= cd -- "$SCRIPT_DIR/.." && pwd)"

usage() {
	cat >&2 <<'EOF'
用法:
  go-modules.sh list
  go-modules.sh resolve <目录>
  go-modules.sh package <exact|recursive> <目录>
  go-modules.sh name <目录>
  go-modules.sh run-target <exact|recursive> <go命令> <目录> [参数...]
  go-modules.sh run-all <go命令> [参数...]
  go-modules.sh build-bin <目录> <输出文件>
  go-modules.sh cover <目录>
  go-modules.sh fmt [目录]
  go-modules.sh tidy [目录]
EOF
}

repo_relative() {
	local path="$1"
	if [[ "$path" == "$ROOT_DIR" ]]; then
		printf '.\n'
	else
		printf '%s\n' "${path#"$ROOT_DIR"/}"
	fi
}

module_dirs() {
	find "$ROOT_DIR" \
		-type d \( -name .git -o -name vendor -o -name bin -o -name node_modules \) -prune -o \
		-type f -name go.mod -print 2>/dev/null |
	while IFS= read -r modfile; do
		repo_relative "$(CDPATH= cd -- "$(dirname -- "$modfile")" && pwd)"
	done | sort
}

absolute_target_dir() {
	local target="${1:-.}"
	local path

	if [[ "$target" == /* ]]; then
		path="$target"
	else
		path="$ROOT_DIR/${target#./}"
	fi

	if [[ ! -e "$path" ]]; then
		printf '目标不存在: %s\n' "$target" >&2
		return 2
	fi

	if [[ -d "$path" ]]; then
		CDPATH= cd -- "$path" && pwd
	else
		CDPATH= cd -- "$(dirname -- "$path")" && pwd
	fi
}

module_root_for_target() {
	local target_dir="$1"
	local current="$target_dir"
	local parent

	while :; do
		if [[ -f "$current/go.mod" ]]; then
			printf '%s\n' "$current"
			return 0
		fi
		[[ "$current" == "$ROOT_DIR" ]] && break
		parent="$(dirname -- "$current")"
		[[ "$parent" == "$current" ]] && break
		current="$parent"
	done

	printf '目标不属于任何 Go module: %s\n' "$target_dir" >&2
	return 2
}

target_context() {
	local target="$1"
	local target_dir module_root rel

	target_dir="$(absolute_target_dir "$target")" || return
	module_root="$(module_root_for_target "$target_dir")" || return

	if [[ "$target_dir" == "$module_root" ]]; then
		rel='.'
	else
		rel="${target_dir#"$module_root"/}"
	fi

	printf '%s\t%s\t%s\n' "$module_root" "$rel" "$(repo_relative "$module_root")"
}

package_for() {
	local mode="$1" target="$2" context module_root rel
	context="$(target_context "$target")" || return
	IFS=$'\t' read -r module_root rel _ <<< "$context"

	case "$mode" in
	exact)
		if [[ "$rel" == '.' ]]; then printf '.\n'; else printf './%s\n' "$rel"; fi
		;;
	recursive)
		if [[ "$rel" == '.' ]]; then printf './...\n'; else printf './%s/...\n' "$rel"; fi
		;;
	*)
		printf '未知 package 模式: %s\n' "$mode" >&2
		return 2
		;;
	esac
}

run_target() {
	local mode="$1" command="$2" target="$3"
	shift 3

	local context module_root package
	context="$(target_context "$target")" || return
	IFS=$'\t' read -r module_root _ _ <<< "$context"
	package="$(package_for "$mode" "$target")" || return

	if [[ "$command" == 'run' ]]; then
		( cd "$module_root" && go run "$package" "$@" )
	else
		( cd "$module_root" && go "$command" "$@" "$package" )
	fi
}

run_all() {
	local command="$1"
	shift
	local module_rel module_root ret=0

	while IFS= read -r module_rel; do
		if [[ "$module_rel" == '.' ]]; then
			module_root="$ROOT_DIR"
		else
			module_root="$ROOT_DIR/$module_rel"
		fi
		printf '\n== module: %s ==\n' "$module_rel" >&2
		if [[ "$command" == 'run' ]]; then
			( cd "$module_root" && go run . "$@" ) || ret=1
		else
			( cd "$module_root" && go "$command" "$@" './...' ) || ret=1
		fi
	done < <(module_dirs)

	return "$ret"
}

format_files() {
	local scope="$1" write="$2" file found=0
	while IFS= read -r -d '' file; do
		found=1
		if [[ "$write" == 'yes' ]]; then
			gofmt -w "$file" || return
		else
			gofmt -l "$file"
		fi
	done < <(find "$scope" -type d \( -name .git -o -name vendor -o -name bin -o -name node_modules \) -prune -o -type f -name '*.go' -print0 2>/dev/null)
	return 0
}

format_command() {
	local target="${1:-}" scope module_rel module_root
	if [[ -n "$target" ]]; then
		scope="$(absolute_target_dir "$target")" || return
		format_files "$scope" yes
		context="$(target_context "$target")" || return
		IFS=$'\t' read -r module_root _ module_rel <<< "$context"
		( cd "$module_root" && go mod tidy )
		return
	fi

	format_files "$ROOT_DIR" yes || return
	while IFS= read -r module_rel; do
		if [[ "$module_rel" == '.' ]]; then module_root="$ROOT_DIR"; else module_root="$ROOT_DIR/$module_rel"; fi
		( cd "$module_root" && go mod tidy ) || return
	done < <(module_dirs)
}

tidy_command() {
	local target="${1:-}" context module_root module_rel
	if [[ -n "$target" ]]; then
		context="$(target_context "$target")" || return
		IFS=$'\t' read -r module_root _ module_rel <<< "$context"
		printf '整理 module: %s\n' "$module_rel"
		( cd "$module_root" && go mod tidy )
		return
	fi

	while IFS= read -r module_rel; do
		if [[ "$module_rel" == '.' ]]; then module_root="$ROOT_DIR"; else module_root="$ROOT_DIR/$module_rel"; fi
		printf '整理 module: %s\n' "$module_rel"
		( cd "$module_root" && go mod tidy ) || return
	done < <(module_dirs)
}

build_bin() {
	local target="$1" output="$2" context module_root package
	context="$(target_context "$target")" || return
	IFS=$'\t' read -r module_root _ _ <<< "$context"
	package="$(package_for exact "$target")" || return
	mkdir -p "$(dirname -- "$output")"
	( cd "$module_root" && go build -o "$output" "$package" )
}

cover_command() {
	local target="$1" context module_root package
	context="$(target_context "$target")" || return
	IFS=$'\t' read -r module_root _ _ <<< "$context"
	package="$(package_for recursive "$target")" || return
	local output="$ROOT_DIR/coverage.out"
	( cd "$module_root" && go test -coverprofile="$output" "$package" ) || return
	go tool cover -html="$output" -o "$ROOT_DIR/coverage.html"
	printf '已生成 %s 和 %s\n' "$ROOT_DIR/coverage.out" "$ROOT_DIR/coverage.html"
}

command="${1:-}"
shift || true
case "$command" in
	list)
		module_dirs
		;;
	resolve)
		[[ $# -eq 1 ]] || { usage; exit 2; }
		context="$(target_context "$1")" || exit
		IFS=$'\t' read -r _ _ module_rel <<< "$context"
		printf '%s\n' "$module_rel"
		;;
	package)
		[[ $# -eq 2 ]] || { usage; exit 2; }
		package_for "$1" "$2"
		;;
	name)
		[[ $# -eq 1 ]] || { usage; exit 2; }
		context="$(target_context "$1")" || exit
		IFS=$'\t' read -r module_root _ _ <<< "$context"
		( cd "$module_root" && go list -f '{{.Name}}' "$(package_for exact "$1")" )
		;;
	run-target)
		[[ $# -ge 3 ]] || { usage; exit 2; }
		run_target "$@"
		;;
	run-all)
		[[ $# -ge 1 ]] || { usage; exit 2; }
		run_all "$@"
		;;
	build-bin)
		[[ $# -eq 2 ]] || { usage; exit 2; }
		build_bin "$1" "$2"
		;;
	cover)
		[[ $# -eq 1 ]] || { usage; exit 2; }
		cover_command "$1"
		;;
	fmt)
		[[ $# -le 1 ]] || { usage; exit 2; }
		format_command "${1:-}"
		;;
	tidy)
		[[ $# -le 1 ]] || { usage; exit 2; }
		tidy_command "${1:-}"
		;;
	*)
		usage
		exit 2
		;;
esac
