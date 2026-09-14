package main

import (
	"fmt"
	"strings"
	"testing"
)

func TestAddInt8_Overflow(t *testing.T) {
	// 这个测试是在"记录"语言行为，而不是在验证业务逻辑。
	// 理解溢出比记住结论重要：int8 只有 8 位，超了就从最小值继续。
	if got := AddInt8(127, 1); got != -128 {
		t.Errorf("127+1 应该溢出为 -128，实际 %d", got)
	}
	if got := AddInt8(-128, -1); got != 127 {
		t.Errorf("-128-1 应该溢出为 127，实际 %d", got)
	}
}

func TestIsEven(t *testing.T) {
	tests := []struct {
		in   int
		want bool
	}{
		{0, true}, {1, false}, {2, true}, {-2, true}, {-3, false}, {100, true},
	}
	for _, tt := range tests {
		if got := IsEven(tt.in); got != tt.want {
			t.Errorf("IsEven(%d) = %t, 想要 %t", tt.in, got, tt.want)
		}
	}
}

func TestGrade(t *testing.T) {
	tests := []struct {
		score int
		want  string
	}{
		{100, "A"}, {90, "A"}, {89, "B"}, {80, "B"}, {79, "C"},
		{70, "C"}, {69, "D"}, {60, "D"}, {59, "F"}, {0, "F"},
		{-1, "非法分数"}, {101, "非法分数"},
	}
	for _, tt := range tests {
		if got := Grade(tt.score); got != tt.want {
			t.Errorf("Grade(%d) = %q, 想要 %q", tt.score, got, tt.want)
		}
	}
}

func TestReverse(t *testing.T) {
	tests := []struct{ in, want string }{
		{"", ""},
		{"a", "a"},
		{"abc", "cba"},
		{"你好世界", "界世好你"}, // 中文必须整体反转
		{"a你b好", "好b你a"}, // 混排
		{"🙂🙃", "🙃🙂"},     // emoji 是 4 字节，按 rune 处理才不会碎
	}
	for _, tt := range tests {
		if got := Reverse(tt.in); got != tt.want {
			t.Errorf("Reverse(%q) = %q, 想要 %q", tt.in, got, tt.want)
		}
	}
}

func TestCountVowels(t *testing.T) {
	tests := []struct {
		in   string
		want int
	}{
		{"", 0},
		{"aeiou", 5},
		{"AEIOU", 5}, // 大小写都要算
		{"Hello", 2},
		{"xyz", 0},
		{"Go语言", 1},
	}
	for _, tt := range tests {
		if got := CountVowels(tt.in); got != tt.want {
			t.Errorf("CountVowels(%q) = %d, 想要 %d", tt.in, got, tt.want)
		}
	}
}

func TestPermission(t *testing.T) {
	var p Permission = PermRead | PermWrite

	if !p.Has(PermRead) || !p.Has(PermWrite) {
		t.Error("应该有读和写权限")
	}
	if p.Has(PermExecute) {
		t.Error("不该有执行权限")
	}
	// Has 要求"同时包含所有指定位"：p 有读+写，所以 读|写 这个组合它是有的
	if !p.Has(PermRead | PermWrite) {
		t.Error("p 同时含读和写，Has(读|写) 应为 true")
	}
	// 只有读权限时，问"有没有写"必须是否
	if PermRead.Has(PermWrite) {
		t.Error("只有读权限时不该报告有写权限")
	}

	// 位清除：去掉写权限
	if got := p &^ PermWrite; got != PermRead {
		t.Errorf("清除写权限后应只剩读，实际 %v", got)
	}
	if PermAll.Has(PermRead) == false {
		t.Error("全部权限应包含读")
	}
}

// String() 实现了 fmt.Stringer，所以 %v 会走我们的逻辑。
// 顺带演示 map 遍历顺序随机 -> 我们必须自己排序才有稳定输出。
func TestPermissionString(t *testing.T) {
	tests := []struct {
		in   Permission
		want string
	}{
		{0, "无权限"},
		{PermRead, "读"},
		{PermWrite, "写"},
		{PermExecute, "执行"},
		{PermRead | PermWrite, "读|写"},
		{PermAll, "读|写|执行"},
	}
	for _, tt := range tests {
		if got := tt.in.String(); got != tt.want {
			t.Errorf("Permission(%d).String() = %q, 想要 %q", tt.in, got, tt.want)
		}
	}
}

func TestSumOfMultiples(t *testing.T) {
	tests := []struct {
		n, k, want int
	}{
		{10, 2, 30}, // 2+4+6+8+10
		{10, 3, 18}, // 3+6+9
		{10, 11, 0}, // 没有倍数
		{0, 2, 0},
		{1, 1, 1},
	}
	for _, tt := range tests {
		if got := SumOfMultiples(tt.n, tt.k); got != tt.want {
			t.Errorf("SumOfMultiples(%d, %d) = %d, 想要 %d", tt.n, tt.k, got, tt.want)
		}
	}
}

func ExampleReverse() {
	fmt.Println(Reverse("Go语言"))
	fmt.Println(strings.ToUpper("go"))
	// Output:
	// 言语oG
	// GO
}
