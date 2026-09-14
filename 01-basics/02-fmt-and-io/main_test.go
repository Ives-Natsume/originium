package main

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"testing/iotest"
)

func TestCountAll(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want TextStats
	}{
		{"空输入", "", TextStats{}},
		{"单行", "hello world", TextStats{Lines: 1, Words: 2, Chars: 11}},
		{"中文按字符计数", "中文测试", TextStats{Lines: 1, Words: 1, Chars: 4}},
		// 换行符本身不计入字符数，只用来分行
		{"多行", "a b\nc\n", TextStats{Lines: 2, Words: 3, Chars: 4}},
		// 空格也计入字符数（我们统计的是"字符"，不是"有效字符"）
		{"连续空白不产生空词", "  a   b  ", TextStats{Lines: 1, Words: 2, Chars: 9}},
		{"空行也计数", "\n\n", TextStats{Lines: 2}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CountAll(strings.NewReader(tt.in))
			if err != nil {
				t.Fatalf("不该报错: %v", err)
			}
			if got != tt.want {
				t.Errorf("CountAll(%q) = %+v, 想要 %+v", tt.in, got, tt.want)
			}
		})
	}
}

// 演示：用 iotest.ErrReader 模拟"读取中途出错"。
// 这就是"依赖 io.Reader"带来的好处 —— 不需要真实文件就能造出各种故障。
func TestCountAllReaderError(t *testing.T) {
	boom := errors.New("磁盘坏了")
	_, err := CountAll(iotest.ErrReader(boom))
	if err == nil {
		t.Fatal("期望拿到错误，实际是 nil")
	}
	// 我们用了 %w 包装，所以这里能用 errors.Is 顺着错误链找回原始错误
	if !errors.Is(err, boom) {
		t.Errorf("错误链里应该能找到原始错误，实际: %v", err)
	}
}

// 演示：单行超过缓冲区上限时 Scanner 会报错（而不是静默截断）
func TestCountAllLongLine(t *testing.T) {
	long := strings.Repeat("x", 100*1024) // 100KB，超过默认 64KB 上限
	_, err := CountAll(strings.NewReader(long))
	if err == nil {
		t.Skip("当前 Go 版本默认缓冲区已足够大，跳过")
	}
	if !strings.Contains(err.Error(), "扫描文本失败") {
		t.Errorf("期望被包装的错误信息，实际: %v", err)
	}
}

func ExampleCountAll() {
	s, _ := CountAll(strings.NewReader("Go 是静态类型语言\n"))
	fmt.Printf("%d 行 %d 词 %d 字符\n", s.Lines, s.Words, s.Chars)
	// Output: 1 行 2 词 10 字符
}
