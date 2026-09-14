// Package main —— 第 1 课：Hello, Go。
//
// 运行: make run p=01-basics/01-hello
// 传参: make run p=01-basics/01-hello Neo
// 测试: make test p=01-basics/01-hello
package main

import (
	"fmt"
	"os"
	"runtime"
	"strings"
)

// Greet 返回一句问候语。
//
// 命名规则：首字母大写 = 导出（包外可见）；小写 = 包内私有。
// 这不是风格建议，是语言规则 —— 编译器和其它包都按这个判断可见性。
func Greet(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		name = "世界"
	}
	return "你好, " + name + "!"
}

func main() {
	// ---- 1. 最基本的输出 ----
	// Println: 各参数之间加空格，末尾加换行
	fmt.Println(Greet(""))
	// Printf: 格式化输出，%s 字符串 %d 整数 %q 带引号的字符串 %v 万能格式
	fmt.Printf("1+1 = %d, %q 的长度是 %d\n", 1+1, "go", len("go"))

	// ---- 2. 命令行参数 ----
	// os.Args[0] 是程序自身的路径，真正的参数从 [1] 开始
	name := "Gopher"
	if len(os.Args) > 1 {
		name = os.Args[1]
	}
	fmt.Printf("收到 %d 个参数: %q\n", len(os.Args)-1, os.Args[1:])
	fmt.Println(Greet(name))

	// ---- 3. 运行环境 ----
	fmt.Printf("Go 版本   = %s\n", runtime.Version())
	fmt.Printf("OS/ARCH   = %s/%s\n", runtime.GOOS, runtime.GOARCH)
	fmt.Printf("逻辑 CPU  = %d\n", runtime.NumCPU())

	// ---- 4. 退出码 ----
	// 正常结束返回 0；os.Exit(1) 表示失败。
	// 脚本和 CI 靠退出码判断成功与否，所以命令行程序的错误要用非 0 退出。
	// 下面这行会立刻终止程序（defer 不会执行），取消注释感受一下：
	// os.Exit(1)
}
