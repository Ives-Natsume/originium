package greeter

import (
	"strings"
	"testing"
)

func TestGreet(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"空串用默认", "", "你好, 世界!"},
		{"普通名字", "Neo", "你好, Neo!"},
		{"裁剪空格", "  Neo  ", "你好, Neo!"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Greet(tt.in); got != tt.want {
				t.Errorf("Greet(%q) = %q, 想要 %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestShout(t *testing.T) {
	// 导出函数走的是 未导出 adorn() -> 导出 Greet() 的调用链，
	// 包外只能通过 Shout 看到合成结果。
	got := Shout("neo")
	if !strings.Contains(got, "NEO") || !strings.HasPrefix(got, "***") {
		t.Errorf("Shout(\"neo\") = %q，应该大写并带装饰", got)
	}
}

// 包内测试（package greeter，不是 greeter_test）可以访问未导出成员。
func TestUnexportedAccessibleInsidePackage(t *testing.T) {
	if version == "" {
		t.Error("包内测试可以直接读未导出的 version")
	}
	if got := adorn("x"); got != "*** x ***" {
		t.Errorf("adorn(\"x\") = %q", got)
	}
}

func TestVersion(t *testing.T) {
	if got := Version(); got != version {
		t.Errorf("Version() = %q, version = %q", got, version)
	}
}

func ExampleGreet() {
	// Example 用包名做前缀时，出现在文档里的位置会更显眼
	Greet("Go")
	Greet("李雷")
	// Output:
}

func ExampleShout() {
	Shout("Go")
	// Output:
}
