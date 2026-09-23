#!/usr/bin/env bash
# ---------------------------------------------------------------------------
# 新建练习骨架
#   scripts/new.sh <目录> app   可运行程序（main.go + main_test.go）
#   scripts/new.sh <目录> lib   库包（含表驱动测试模板）
#   scripts/new.sh <目录> project 独立 Go 项目（含 go.mod）
# ---------------------------------------------------------------------------
set -euo pipefail

dir="${1:-}"
mode="${2:-app}"

if [[ -z "$dir" ]]; then
	printf '用法: scripts/new.sh <目录> [app|lib|project]\n' >&2
	exit 2
fi

case "$mode" in
	app|lib|project) ;;
	*)
		printf '未知模式: %s（可选: app、lib、project）\n' "$mode" >&2
		exit 2
		;;
esac

if [[ -e "$dir" ]]; then
	printf '\033[31m目录已存在: %s\033[0m\n' "$dir" >&2
	exit 1
fi

name="$(basename "$dir")"
# main.go 的 package 名不能带连字符/数字开头
pkg="$(printf '%s' "$name" | tr -c 'a-zA-Z0-9_' '_' | tr -s '_')"
[[ "$pkg" =~ ^[0-9] ]] && pkg="p_$pkg"

root_module="$(go list -m -f '{{.Path}}' 2>/dev/null || true)"
if [[ -z "$root_module" ]]; then
	printf '无法读取根模块路径，请在包含 go.mod 的仓库内执行。\n' >&2
	exit 2
fi

mkdir -p "$dir"

touch_go() { printf '%s\n' "$2" > "$dir/$1"; }

if [[ "$mode" == "project" ]]; then
	module_path="$root_module/${dir#./}"

	touch_go "go.mod" "$(cat <<EOF
module $module_path

go 1.27.1
EOF
)"

	touch_go "main.go" "$(cat <<EOF
package main

import "fmt"

func main() {
	fmt.Println("hello from $dir")
}
EOF
)"

	touch_go "main_test.go" "$(cat <<EOF
package main

import "testing"

func TestSmoke(t *testing.T) {
	if got := 1 + 1; got != 2 {
		t.Fatalf("1+1 = %d, 想要 2", got)
	}
}
EOF
)"

	printf '%s\n' "$(cat <<EOF
# $name

这是一个独立的 Go module，可以单独添加第三方依赖并运行。

运行：

    make run p=$dir

测试：

    make test p=$dir

该项目拥有自己的 \`go.mod\`，不会把依赖写入仓库根模块。
EOF
)" > "$dir/README.md"

	printf '\033[32m✓\033[0m 已创建独立项目 %s\n' "$dir"
	printf '  模块: %s\n' "$module_path"
	printf '  运行: make run p=%s\n' "$dir"
	printf '  测试: make test p=%s\n' "$dir"
elif [[ "$mode" == "lib" ]]; then
	touch_go "$name.go" "$(cat <<EOF
// Package $pkg 演示 …（一句话说明这个包做什么）。
package $pkg

// Add 返回 a+b。
func Add(a, b int) int {
	return a + b
}
EOF
)"

	touch_go "${name}_test.go" "$(cat <<EOF
package $pkg

import "testing"

// 表驱动测试：Go 里最常用的测试写法。
func TestAdd(t *testing.T) {
	tests := []struct {
		name string
		a, b int
		want int
	}{
		{"正数", 1, 2, 3},
		{"含零", 0, 5, 5},
		{"负数", -1, 1, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Add(tt.a, tt.b); got != tt.want {
				t.Errorf("Add(%d, %d) = %d, 想要 %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}
EOF
)"
	printf '\033[32m✓\033[0m 已创建库包 %s\n' "$dir"
	printf '  导入方式: import "%s/%s"\n' "$root_module" "${dir#./}"
else
	touch_go "main.go" "$(cat <<EOF
package main

import "fmt"

func main() {
	fmt.Println("hello from $dir")
}
EOF
)"

	touch_go "main_test.go" "$(cat <<EOF
package main

import "testing"

func TestSmoke(t *testing.T) {
	// 冒烟测试占位：换成真正要验证的逻辑
	if got := 1 + 1; got != 2 {
		t.Fatalf("1+1 = %d, 想要 2", got)
	}
}
EOF
)"
	printf '\033[32m✓\033[0m 已创建可运行包 %s\n' "$dir"
	printf '  运行: make run p=%s\n' "$dir"
	printf '  测试: make test p=%s\n' "$dir"
fi