// Package main —— 第 3 课：类型、常量与控制流。
//
// 运行: make run p=01-basics/03-types-and-control
package main

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"unsafe"
)

// ---------------- 类型转换与数值边界 ----------------

// AddInt8 演示溢出：int8 范围是 -128..127，加过头会"绕回去"。
// 这种 bug 在生产里很隐蔽，所以选类型时要想清楚范围。
func AddInt8(a, b int8) int8 { return a + b }

// IsEven 判断偶数。注意负数取余在 Go 里是"截断"的：-3 % 2 == -1。
func IsEven(n int) bool { return n%2 == 0 }

// Grade 根据分数返回等级。演示 switch 的多种形态。
func Grade(score int) string {
	switch {
	case score < 0 || score > 100:
		return "非法分数"
	case score >= 90:
		return "A"
	case score >= 80:
		return "B"
	case score >= 70:
		return "C"
	case score >= 60:
		return "D"
	default:
		return "F"
	}
}

// Reverse 反转字符串。
//
// 关键点：Go 的 string 是**不可变的 UTF-8 字节序列**，
// 中文一个字占 3 字节。按字节反转会把中文切碎成乱码，
// 所以必须先转成 []rune（码点）再反转。
func Reverse(s string) string {
	rs := []rune(s)
	for i, j := 0, len(rs)-1; i < j; i, j = i+1, j-1 {
		rs[i], rs[j] = rs[j], rs[i]
	}
	return string(rs)
}

// CountVowels 统计元音字母，演示 range over string 的行为。
func CountVowels(s string) int {
	n := 0
	for _, r := range strings.ToLower(s) { // range 字符串时 r 是 rune，不是 byte
		if strings.ContainsRune("aeiou", r) {
			n++
		}
	}
	return n
}

// ---------------- 常量与 iota ----------------

// Permission 用一个字节表示权限位，是 iota 最经典的用法。
type Permission uint8

const (
	PermRead    Permission = 1 << iota // 1 << 0 = 1
	PermWrite                          // 1 << 1 = 2
	PermExecute                        // 1 << 2 = 4
	PermAll     = PermRead | PermWrite | PermExecute
)

// Has 检查是否包含某个权限位。
func (p Permission) Has(want Permission) bool { return p&want == want }

// String 让 Permission 打印出来是人能读的，而不是数字。
// 实现了 fmt.Stringer 接口，fmt 就会自动调用它。
func (p Permission) String() string {
	if p == 0 {
		return "无权限"
	}
	var parts []string
	for bit, name := range map[Permission]string{
		PermRead: "读", PermWrite: "写", PermExecute: "执行",
	} {
		if p.Has(bit) {
			parts = append(parts, name)
		}
	}
	// map 遍历顺序随机，排一下保证输出稳定
	order := []string{"读", "写", "执行"}
	var sorted []string
	for _, o := range order {
		for _, got := range parts {
			if got == o {
				sorted = append(sorted, got)
			}
		}
	}
	return strings.Join(sorted, "|")
}

// SumOfMultiples 计算 1..n 中能被 k 整除的数之和。
// 演示 for 的几种写法和带标签的 break。
func SumOfMultiples(n, k int) int {
	sum := 0
outer: // 标签：可以从内层循环直接跳出外层
	for i := 1; i <= n; i++ {
		if i > 1000 {
			break outer // 保护性上限
		}
		if i%k != 0 {
			continue
		}
		sum += i
	}
	return sum
}

func main() {
	fmt.Println("---- 类型的尺寸与零值 ----")
	var (
		i   int
		i8  int8
		i64 int64
		f32 float32
		f64 float64
		s   string
		b   bool
		p   *int
	)
	fmt.Printf("int=%d位 零值=%d\n", unsafe.Sizeof(i)*8, i)
	fmt.Printf("int8=%d位 范围=%d..%d 零值=%d\n", unsafe.Sizeof(i8)*8, math.MinInt8, math.MaxInt8, i8)
	fmt.Printf("int64=%d位 零值=%d\n", unsafe.Sizeof(i64)*8, i64)
	fmt.Printf("float32=%d位 零值=%v\n", unsafe.Sizeof(f32)*8, f32)
	fmt.Printf("float64=%d位 零值=%v\n", unsafe.Sizeof(f64)*8, f64)
	fmt.Printf("string 零值=%q  len=%d\n", s, len(s))
	fmt.Printf("bool 零值=%v\n", b)
	fmt.Printf("指针零值=%v（nil，解引用会 panic）\n", p)

	fmt.Println("\n---- 溢出：Go 不会替你检查 ----")
	fmt.Printf("int8: 127 + 1 = %d   ← 绕回负数了\n", AddInt8(127, 1))
	fmt.Printf("int8: -128 - 1 = %d  ← 绕回正数了\n", AddInt8(-128, -1))
	fmt.Println("所以：涉及金钱用 int64 存最小单位，涉及范围先想清楚再选类型")

	fmt.Println("\n---- 浮点精度 ----")
	fmt.Printf("0.1 + 0.2 = %.20f\n", 0.1+0.2)
	// 注意：这个字面量比较在同一表达式里会被编译器按常量计算并优化掉，
	// 看起来"相等"其实不可靠。要看真实行为得放进变量。
	x, y := 0.1, 0.2
	fmt.Printf("变量相加 0.1+0.2 == 0.3 ? %t\n", x+y == 0.3)
	fmt.Printf("用误差范围比较: |diff| < 1e-9 ? %t\n", math.Abs((x+y)-0.3) < 1e-9)
	fmt.Println("结论：浮点数永远不要用 == 比较，要么用误差范围，要么改用整数存最小单位")
	fmt.Printf("float32 累加 1000 次 0.1 = %v（误差被放大到肉眼可见）\n", func() float32 {
		var acc float32
		for range 1000 { // Go 1.22+ 可以 range 一个整数
			acc += 0.1
		}
		return acc
	}())

	fmt.Println("\n---- 字符串 / 字节 / rune ----")
	text := "Go语言"
	fmt.Printf("%q 字节数=%d 字符数=%d\n", text, len(text), len([]rune(text)))
	for i, r := range text {
		fmt.Printf("  字节下标=%-2d rune=%q 码点=U+%04X\n", i, r, r)
	}
	fmt.Printf("按字节反转（错的）: %q\n", reverseBytes(text))
	fmt.Printf("按 rune 反转（对的）: %q\n", Reverse(text))

	fmt.Println("\n---- 常量与 iota 位运算 ----")
	var perm Permission = PermRead | PermExecute
	fmt.Printf("权限 = %v  数值 = %d\n", perm, perm)
	fmt.Printf("有读权限? %t   有写权限? %t\n", perm.Has(PermRead), perm.Has(PermWrite))
	fmt.Printf("全部权限 = %v (%d)\n", PermAll, PermAll)
	fmt.Printf("去掉写权限 = %v\n", PermAll&^PermWrite) // &^ 是"位清除"，Go 独有的运算符

	fmt.Println("\n---- 控制流 ----")
	for _, sc := range []int{95, 85, 72, 61, 30, 120} {
		fmt.Printf("分数 %3d -> %s\n", sc, Grade(sc))
	}
	fmt.Printf("1..100 内 7 的倍数之和 = %d\n", SumOfMultiples(100, 7))
	fmt.Printf("元音字母个数 = %d\n", CountVowels("Hello, Gopher!"))
	fmt.Printf("数字转字符串: %q\n", strconv.Itoa(42))
	if n, err := strconv.Atoi("123"); err == nil {
		fmt.Printf("字符串转数字: %d\n", n)
	}
}

// reverseBytes 是错误示范：按字节反转会切碎多字节字符。
func reverseBytes(s string) string {
	bs := []byte(s)
	for i, j := 0, len(bs)-1; i < j; i, j = i+1, j-1 {
		bs[i], bs[j] = bs[j], bs[i]
	}
	return string(bs)
}
