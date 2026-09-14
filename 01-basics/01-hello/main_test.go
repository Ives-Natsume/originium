package main

import (
	"fmt"
	"testing"
)

// 表驱动测试（table-driven test）：Go 里最主流的测试写法。
// 把"输入 -> 期望输出"列成表，循环跑。加用例只要加一行。
func TestGreet(t *testing.T) {
	tests := []struct {
		name string // 用例名字，会显示在失败信息里
		in   string
		want string
	}{
		{"空串用默认值", "", "你好, 世界!"},
		{"常见名字", "Neo", "你好, Neo!"},
		{"两端空格被裁剪", "  Neo  ", "你好, Neo!"},
		{"全空格视为空", "   ", "你好, 世界!"},
		{"中文原样保留", "李雷", "你好, 李雷!"},
	}

	for _, tt := range tests {
		// t.Run 创建子测试：失败时能精确定位是哪个用例
		t.Run(tt.name, func(t *testing.T) {
			if got := Greet(tt.in); got != tt.want {
				t.Errorf("Greet(%q) = %q, 想要 %q", tt.in, got, tt.want)
			}
		})
	}
}

// Example 测试：注释里的 "// Output:" 就是期望的标准输出。
// go test 会自动运行并比对，同时它会出现在文档里（go doc）。
func ExampleGreet() {
	fmt.Println(Greet("Go"))
	// Output: 你好, Go!
}

// Benchmark 基准测试：go test -bench=. 才会跑。
// make bench p=01-basics/01-hello
func BenchmarkGreet(b *testing.B) {
	for b.Loop() { // Go 1.24+ 推荐写法，自动处理计时与迭代次数
		_ = Greet("Neo")
	}
}
