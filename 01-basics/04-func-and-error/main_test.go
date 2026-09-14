package main

import (
	"errors"
	"fmt"
	"testing"
)

func TestDiv(t *testing.T) {
	tests := []struct {
		name    string
		a, b    int
		want    int
		wantErr error
	}{
		{"正常整除", 10, 2, 5, nil},
		{"除不尽取整", 7, 2, 3, nil},
		{"负数", -10, 2, -5, nil},
		{"除零返回哨兵错误", 10, 0, 0, ErrDivByZero},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Div(tt.a, tt.b)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("错误 = %v, 想要 %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("Div(%d,%d) = %d, 想要 %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

// 关键测试：包装之后，errors.Is 依然能找到原始哨兵错误。
func TestSafeDiv_ErrorChain(t *testing.T) {
	_, err := SafeDiv(10, 0)
	if err == nil {
		t.Fatal("期望报错")
	}

	// 1. 错误信息里有上下文
	const wantMsg = "计算 10/0 失败"
	if got := err.Error(); len(got) < len(wantMsg) || got[:len(wantMsg)] != wantMsg {
		t.Errorf("错误信息应以 %q 开头，实际 %q", wantMsg, got)
	}

	// 2. errors.Is 能穿透包装找到根因
	if !errors.Is(err, ErrDivByZero) {
		t.Error("错误链里应该能找到 ErrDivByZero")
	}

	// 3. 错误信息不会因为包装而丢字
	if errors.Is(err, ErrTooShort) {
		t.Error("不该匹配到无关的错误")
	}
}

func TestValidateName(t *testing.T) {
	t.Run("合法名字不报错", func(t *testing.T) {
		for _, n := range []string{"Neo", "李雷", "ab", "  Neo  "} {
			if err := ValidateName(n); err != nil {
				t.Errorf("ValidateName(%q) 不该报错: %v", n, err)
			}
		}
	})

	t.Run("空名字报错且能取到字段名", func(t *testing.T) {
		err := ValidateName("")
		var ve *ValidationError
		// errors.As 会沿着错误链找类型
		if !errors.As(err, &ve) {
			t.Fatalf("期望 *ValidationError，实际 %T: %v", err, err)
		}
		if ve.Field != "name" {
			t.Errorf("Field = %q, 想要 name", ve.Field)
		}
	})

	t.Run("单字符触发 ErrTooShort", func(t *testing.T) {
		err := ValidateName("A")
		if !errors.Is(err, ErrTooShort) {
			t.Errorf("期望 ErrTooShort，实际 %v", err)
		}
		// 同时也能 As 出结构化字段 —— 两者不冲突
		var ve *ValidationError
		if !errors.As(err, &ve) {
			t.Fatal("也应该能 As 出 *ValidationError")
		}
		if ve.Err != ErrTooShort {
			t.Errorf("ve.Err = %v, 想要 ErrTooShort", ve.Err)
		}
	})
}

func TestSum(t *testing.T) {
	tests := []struct {
		name string
		in   []int
		want int
	}{
		{"无参数", nil, 0},
		{"单个", []int{5}, 5},
		{"多个", []int{1, 2, 3}, 6},
		{"含负数", []int{-1, 1}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Sum(tt.in...); got != tt.want {
				t.Errorf("Sum(%v) = %d, 想要 %d", tt.in, got, tt.want)
			}
		})
	}
}

func TestStats(t *testing.T) {
	mn, mx, sm, ok := Stats(3, 1, 4, 1, 5)
	if !ok {
		t.Fatal("有输入时 ok 应为 true")
	}
	if mn != 1 || mx != 5 || sm != 14 {
		t.Errorf("Stats = (%d,%d,%d), 想要 (1,5,14)", mn, mx, sm)
	}

	if _, _, _, ok := Stats(); ok {
		t.Error("无输入时 ok 应为 false")
	}
}

// 闭包状态是独立的：这是很多人第一次接触闭包时的疑惑点。
func TestCounter_IndependentState(t *testing.T) {
	a := Counter()
	b := Counter()

	if got := a(); got != 1 {
		t.Errorf("a() = %d, 想要 1", got)
	}
	if got := a(); got != 2 {
		t.Errorf("a() = %d, 想要 2", got)
	}
	// b 是全新的计数器，不受 a 影响
	if got := b(); got != 1 {
		t.Errorf("b() = %d, 想要 1（两个闭包状态互相独立）", got)
	}
}

func TestAccumulator(t *testing.T) {
	add, reset := Accumulator(100)

	if got := add(10); got != 110 {
		t.Errorf("add(10) = %d, 想要 110", got)
	}
	if got := add(20); got != 130 {
		t.Errorf("add(20) = %d, 想要 130", got)
	}
	reset()
	if got := add(0); got != 100 {
		t.Errorf("reset 后 add(0) = %d, 想要 100", got)
	}
}

func TestDeferOrder(t *testing.T) {
	got := DeferOrder()
	want := []string{"3", "2", "1"} // LIFO
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Errorf("DeferOrder() = %v, 想要 %v", got, want)
	}
}

func TestDeferArgs(t *testing.T) {
	got := DeferArgs()
	if len(got) != 2 {
		t.Fatalf("期望 2 条日志，实际 %v", got)
	}
	// 顺序很重要：函数体里的 append 先执行，defer 里的后执行
	if got[0] != "函数结束时 x=100" {
		t.Errorf("函数体先执行，实际 %q", got[0])
	}
	// defer 的实参在 defer 那一刻就求值了，所以捕获到的是 1 而不是 100
	if got[1] != "捕获时 v=1" {
		t.Errorf("defer 的实参应该在 defer 那一刻求值，实际 %q", got[1])
	}
}

func TestNamedReturn(t *testing.T) {
	if got := NamedReturn(); got != 10 {
		t.Errorf("NamedReturn() = %d, 想要 10（5 被 defer 翻倍）", got)
	}
}

func TestWithCleanup(t *testing.T) {
	t.Run("正常路径", func(t *testing.T) {
		got, err := WithCleanup(func() (string, error) { return "ok", nil })
		if err != nil || got != "ok" {
			t.Errorf("got=%q err=%v", got, err)
		}
	})

	t.Run("出错路径下 defer 依然执行", func(t *testing.T) {
		boom := errors.New("干活失败")
		_, err := WithCleanup(func() (string, error) { return "", boom })
		if !errors.Is(err, boom) {
			t.Errorf("err = %v, 想要 %v", err, boom)
		}
		// "关闭"那行由 defer 打印，说明即使出错资源也被释放了
	})
}

func TestMustParseInt(t *testing.T) {
	if got := MustParseInt("42"); got != 42 {
		t.Errorf("MustParseInt(\"42\") = %d", got)
	}

	defer func() {
		r := recover()
		if r == nil {
			t.Error("非法输入应该 panic")
		}
	}()
	MustParseInt("abc")
}

func TestSafeRun(t *testing.T) {
	tests := []struct {
		name    string
		fn      func()
		wantErr bool
	}{
		{"正常函数", func() {}, false},
		{"手动 panic", func() { panic("boom") }, true},
		{"切片越界", func() { _ = []int{}[3] }, true},
		{"nil 解引用", func() { var p *int; _ = *p }, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := SafeRun(tt.fn)
			if (err != nil) != tt.wantErr {
				t.Errorf("SafeRun err = %v, wantErr = %t", err, tt.wantErr)
			}
		})
	}
}

func ExampleSafeDiv() {
	_, err := SafeDiv(10, 0)
	fmt.Println(err)
	fmt.Println(errors.Is(err, ErrDivByZero))
	// Output:
	// 计算 10/0 失败: 除数不能为零
	// true
}
