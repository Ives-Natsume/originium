package app

import (
	"strings"
	"testing"
)

func TestWelcome(t *testing.T) {
	t.Run("正常流程", func(t *testing.T) {
		got, err := Welcome("Neo")
		if err != nil {
			t.Fatalf("不该报错: %v", err)
		}
		if !strings.HasPrefix(got, "你好, Neo!") {
			t.Errorf("缺少问候语前缀: %q", got)
		}
		if !strings.Contains(got, "[token=") {
			t.Errorf("缺少令牌: %q", got)
		}
	})

	t.Run("空名字被拦截", func(t *testing.T) {
		if _, err := Welcome("  "); err == nil {
			t.Error("空名字应该报错")
		}
	})

	t.Run("每次令牌都不同", func(t *testing.T) {
		a, _ := Welcome("Neo")
		b, _ := Welcome("Neo")
		if a == b {
			t.Error("两次调用应该生成不同令牌")
		}
	})
}

func TestDescribe(t *testing.T) {
	lines := Describe()
	if len(lines) != 3 {
		t.Fatalf("期望 3 行，实际 %d: %v", len(lines), lines)
	}
	for i, l := range lines {
		if l == "" {
			t.Errorf("第 %d 行为空", i)
		}
	}
}
