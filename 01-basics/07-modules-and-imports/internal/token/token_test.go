package token

import (
	"encoding/hex"
	"errors"
	"testing"
)

func TestSign(t *testing.T) {
	// 已知答案测试：SHA-256("abc") 是公开固定值，可以用来验证实现没写错
	got, err := Sign("abc")
	if err != nil {
		t.Fatalf("不该报错: %v", err)
	}
	const want = "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad"
	if got != want {
		t.Errorf("Sign(\"abc\") = %q, 想要 %q", got, want)
	}

	t.Run("相同输入结果相同", func(t *testing.T) {
		a, _ := Sign("hello")
		b, _ := Sign("hello")
		if a != b {
			t.Error("纯函数应该对相同输入返回相同输出")
		}
	})

	t.Run("不同输入结果不同", func(t *testing.T) {
		a, _ := Sign("hello")
		b, _ := Sign("hellp") // 只差一个字符
		if a == b {
			t.Error("不同输入不该产生相同摘要")
		}
	})

	t.Run("空输入报错", func(t *testing.T) {
		if _, err := Sign(""); !errors.Is(err, ErrEmpty) {
			t.Errorf("err = %v, 想要 ErrEmpty", err)
		}
	})
}

func TestMustSign(t *testing.T) {
	if got := MustSign("abc"); len(got) != 64 {
		t.Errorf("摘要长度应为 64，实际 %d", len(got))
	}

	defer func() {
		if r := recover(); r == nil {
			t.Error("空输入应该 panic")
		}
	}()
	MustSign("")
}

func TestRandom(t *testing.T) {
	t.Run("长度正确且是合法十六进制", func(t *testing.T) {
		got, err := Random(16)
		if err != nil {
			t.Fatalf("不该报错: %v", err)
		}
		if len(got) != 32 { // 16 字节 -> 32 个十六进制字符
			t.Errorf("长度 = %d, 想要 32", len(got))
		}
		if _, err := hex.DecodeString(got); err != nil {
			t.Errorf("结果不是合法十六进制: %v", err)
		}
	})

	t.Run("两次调用结果不同", func(t *testing.T) {
		a, _ := Random(16)
		b, _ := Random(16)
		if a == b {
			t.Error("随机值不该重复（概率低到可以认为必然不同）")
		}
	})

	t.Run("非法长度报错", func(t *testing.T) {
		for _, n := range []int{0, -1} {
			if _, err := Random(n); err == nil {
				t.Errorf("Random(%d) 应该报错", n)
			}
		}
	})
}
