package main

import (
	"os"
	"strings"
	"testing"

	"github.com/Ives-Natsume/originium/01-basics/07-modules-and-imports/app"
	"github.com/Ives-Natsume/originium/01-basics/07-modules-and-imports/greeter"
	"github.com/Ives-Natsume/originium/01-basics/07-modules-and-imports/internal/token"
)

// 集成测试：这个文件同时用了三个包，验证它们能一起工作。
// 放在 main 包里，所以能同时导入 app / greeter / internal/token。
func TestPackagesWorkTogether(t *testing.T) {
	msg, err := app.Welcome("Neo")
	if err != nil {
		t.Fatalf("Welcome 失败: %v", err)
	}

	// app.Welcome 内部调用了 greeter.Greet，所以消息里应该含问候语
	if !strings.Contains(msg, greeter.Greet("Neo")) {
		t.Errorf("消息里应包含 greeter.Greet 的结果: %q", msg)
	}
}

func TestInternalPackageIsVisibleHere(t *testing.T) {
	// 这个测试能编译通过，本身就是断言：
	// 本包位于 internal 的父目录树内，所以可以导入 token。
	// 如果把这个测试文件搬到 01-basics/01-hello，编译就会失败。
	if got := token.MustSign("x"); len(got) != 64 {
		t.Errorf("摘要长度 = %d, 想要 64", len(got))
	}
}

func TestRun(t *testing.T) {
	t.Run("默认不报错", func(t *testing.T) {
		old := os.Args
		t.Cleanup(func() { os.Args = old })
		os.Args = []string{"prog"}
		if err := run(); err != nil {
			t.Errorf("不该报错: %v", err)
		}
	})

	t.Run("传入 fail 时报错", func(t *testing.T) {
		old := os.Args
		t.Cleanup(func() { os.Args = old })
		os.Args = []string{"prog", "fail"}
		if err := run(); err == nil {
			t.Error("应该报错")
		}
	})
}
