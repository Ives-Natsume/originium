// Package main —— 第 2 课：格式化输出与输入输出流。
//
// 运行: make run p=01-basics/02-fmt-and-io
package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

// TextStats 描述一段文本的统计结果。
type TextStats struct {
	Lines int // 行数
	Words int // 词数（按空白切分）
	Chars int // 字符数（按 rune 计，中文算 1 个）
}

// CountAll 从任意 io.Reader 读取内容并统计。
//
// 这是 Go 里非常重要的一个习惯：**依赖接口而不是具体类型**。
// 参数写 io.Reader，那么文件、网络连接、字符串、压缩流……都能传进来，
// 单元测试里也就不用真去建文件。
func CountAll(r io.Reader) (TextStats, error) {
	var s TextStats
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := sc.Text()
		s.Lines++
		s.Chars += len([]rune(line)) // 转成 rune 切片才是"字符数"
		s.Words += len(strings.Fields(line))
	}
	// Scanner 的错误必须显式检查：一行超过 64KB 就会报错，
	// 忘了检查就会静默丢数据。这是真实项目里常见的坑。
	if err := sc.Err(); err != nil {
		return s, fmt.Errorf("扫描文本失败: %w", err)
	}
	return s, nil
}

func main() {
	// ================= 1. fmt 的常用动词 =================
	type user struct {
		Name string
		Age  int
	}
	u := user{Name: "Neo", Age: 30}

	// 左边一列用 %-6s 把"格式串本身"当字符串打出来，右边是实际结果。
	// 注意：格式串是运行时数据，所以要用变量/字面量传进去，
	// 不能把 %v 直接写在格式串里 —— 那样它就变成动词了。
	fmt.Println("---- fmt 动词 ----")
	fmt.Printf("%-6s 默认格式          %v\n", "%v", u)
	fmt.Printf("%-6s 带字段名          %+v\n", "%+v", u)
	fmt.Printf("%-6s 带类型信息        %#v\n", "%#v", u)
	fmt.Printf("%-6s 类型              %T\n", "%T", u)
	fmt.Printf("%-6s 带引号字符串      %q\n", "%q", "go\n")
	fmt.Printf("%-6s 整数              %d\n", "%d", 42)
	fmt.Printf("%-6s 宽度 5 右对齐      %5d|\n", "%5d", 42)
	fmt.Printf("%-6s 宽度 5 左对齐      %-5d|\n", "%-5d", 42)
	fmt.Printf("%-6s 补零              %05d\n", "%05d", 42)
	fmt.Printf("%-6s 两位小数          %.2f\n", "%.2f", 3.14159)
	fmt.Printf("%-6s 宽度 + 精度        %8.2f|\n", "%8.2f", 3.14159)
	fmt.Printf("%-6s 布尔              %t\n", "%t", true)
	fmt.Printf("%-6s 十六进制          %x / %X\n", "%x/%X", 255, 255)
	fmt.Printf("%-6s 字符              %c\n", "%c", '中')
	fmt.Printf("%-6s 指针地址          %p\n", "%p", &u)
	fmt.Printf("%-6s 百分号本身        百分之 100%%\n", "%%")

	// ================= 2. 拼字符串：Builder 比 += 快得多 =================
	// 用 Sprintf 做对齐，输出一张小表格
	fmt.Println("\n---- 对齐表格 ----")
	rows := []struct {
		name  string
		score float64
	}{
		{"Neo", 91.5},
		{"李雷", 88},
		{"Alexander", 100},
	}
	var b strings.Builder // 反复拼接时用 Builder，避免每次 += 都重新分配内存
	fmt.Fprintf(&b, "%-12s %8s\n", "姓名", "得分")
	fmt.Fprintf(&b, "%s\n", strings.Repeat("-", 21))
	for _, r := range rows {
		// 注意：%-12s 按字节算宽度，中文占 3 字节会看起来不齐，
		// 真实项目里要么用 golang.org/x/text 的宽度计算，要么别混排。
		fmt.Fprintf(&b, "%-12s %8.1f\n", r.name, r.score)
	}
	fmt.Print(b.String())

	// ================= 3. 从 Reader 读：测试里就不用建文件 =================
	fmt.Println("\n---- 读取并统计 ----")
	text := "Go 语言爱好者\n第二行 有 五个词\n"
	stats, err := CountAll(strings.NewReader(text))
	if err != nil {
		// 真实程序这里通常是：记日志 + 返回错误给上层，而不是直接退出
		fmt.Fprintln(os.Stderr, "出错了:", err)
		os.Exit(1)
	}
	fmt.Printf("行=%d 词=%d 字符=%d\n", stats.Lines, stats.Words, stats.Chars)

	// ================= 4. 逐行读取标准输入 =================
	// 取消下面注释后，`make run p=01-basics/02-fmt-and-io` 会等你输入，
	// 输入几行文字按 Ctrl-D 结束。
	fmt.Println("\n---- 逐行读取标准输入（示例代码已注释，可解开体验）----")
	scanStdin()

	// ================= 5. 输出到不同目标 =================
	// fmt.Fprintln 的第一个参数是 io.Writer：
	// os.Stdout / os.Stderr / 文件 / 网络连接 / bytes.Buffer 都可以。
	fmt.Fprintln(os.Stderr, "[stderr] 这行走的是标准错误流，方便和正常输出分开重定向")
}

// scanStdin 演示标准输入读取。
func scanStdin() {
	sc := bufio.NewScanner(os.Stdin)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024) // 放开单行上限到 1MB
	n := 0
	for sc.Scan() {
		n++
		fmt.Printf("%d: %s\n", n, sc.Text())
	}
	if err := sc.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "读取失败:", err)
	}
}
