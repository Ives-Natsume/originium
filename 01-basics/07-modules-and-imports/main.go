// Package main —— 第 7 课：模块、包与导入。
//
// 运行: make run p=01-basics/07-modules-and-imports
package main

import (
	"errors"
	"fmt"
	"log"
	"os"

	// 导入路径 = go.mod 里的 module 名 + 目录的相对路径
	"github.com/Ives-Natsume/originium/01-basics/07-modules-and-imports/app"
	"github.com/Ives-Natsume/originium/01-basics/07-modules-and-imports/greeter"
	"github.com/Ives-Natsume/originium/01-basics/07-modules-and-imports/internal/token"

	// 下面是各种"有特殊写法"的导入，用不到就先别看：

	// 1) 起别名（包名和目录名不一致，或名字太长/易冲突时用）
	strutil "strings"

	// 2) 空导入：只为了触发被导入包里的 init()，不直接使用它。
	//    典型场景：database/sql 的驱动注册 —— import _ "github.com/go-sql-driver/mysql"
	//    第 4 章会真正用到这个技巧。
	_ "os/exec"
	// 3) 点导入：把包内标识符直接注入当前命名空间，写 Println 而不是 fmt.Println。
	//    **强烈不推荐**，因为它让"这个名字从哪来的"变得无从查证，也容易撞名。
	//    这里只做演示：. "fmt"
)

func main() {
	fmt.Println("===== 1. 跨包调用 =====")
	fmt.Printf("greeter.Greet(\"Neo\")   = %q\n", greeter.Greet("Neo"))
	fmt.Printf("greeter.Shout(\"neo\")   = %q\n", greeter.Shout("neo"))
	fmt.Printf("greeter.Version()       = %q\n", greeter.Version())
	fmt.Printf("greeter.DefaultGreeting = %q\n", greeter.DefaultGreeting)

	// 想试编译错误？取消下面这行的注释，然后 `go build ./...`：
	// fmt.Println(greeter.version)        // ✗ 未导出，包外不可见
	// fmt.Println(greeter.adorn("x"))     // ✗ 未导出函数，同样不可见
	fmt.Println("（greeter.version / greeter.adorn 在这个包里看不见 —— 这就是导出规则）")

	fmt.Println("\n===== 2. internal 包：限定可见范围 =====")
	// token 位于 .../07-modules-and-imports/internal/token
	// 所以本包（同目录树内）可以导入；其他章节的代码导入会编译失败：
	//   import "github.com/.../01-basics/07-modules-and-imports/internal/token"  ← 在 02-concurrency 里会报错
	sum := token.MustSign("abc")
	fmt.Printf("token.MustSign(\"abc\")  = %s…（前 16 位）\n", sum[:16])
	if r, err := token.Random(8); err == nil {
		fmt.Printf("token.Random(8)        = %s\n", r)
	}

	fmt.Println("\n===== 3. 包依赖分层 =====")
	fmt.Println("依赖方向: main -> app -> {greeter, internal/token}")
	for _, line := range app.Describe() {
		fmt.Printf("  %s\n", line)
	}

	fmt.Println("\n===== 4. app 包的业务流程 =====")
	for _, name := range []string{"Neo", "李雷", ""} {
		msg, err := app.Welcome(name)
		if err != nil {
			fmt.Printf("  name=%-6q ✗ %v\n", name, err)
			continue
		}
		fmt.Printf("  name=%-6q ✓ %s\n", name, msg)
	}

	// 把底层错误翻译成上层错误，同时保留链路
	_, err := app.Welcome("")
	fmt.Printf("错误链: errors.Is(err, token.ErrEmpty) = %t\n", errors.Is(err, token.ErrEmpty))
	fmt.Println("（这里是新错误，不是包装的，所以 false。要保留链路就得用 %w）")

	fmt.Println("\n===== 5. 导入别名与标准库重名 =====")
	// strutil 是 "strings" 的别名
	fmt.Printf("strutil.ToUpper(\"go\") = %q\n", strutil.ToUpper("go"))

	fmt.Println("\n===== 6. go.mod 里写了什么 =====")
	fmt.Println("  module  <路径>     ← 所有内部导入路径的前缀")
	fmt.Println("  go      <版本>     ← 语言特性开关（别乱改小版本）")
	fmt.Println("  require <依赖>     ← 直接依赖；间接依赖会被标 // indirect")
	fmt.Println("  replace <旧> => <新> ← 本地开发时把依赖指向本地目录，非常有用")
	fmt.Println("  tool / toolchain   ← 固定工具版本，团队协作时保证一致")
	fmt.Println("  用 make deps 会自动整理 go.mod 和 go.sum")

	fmt.Println("\n===== 7. 日志与退出码（命令行程序的基本要求）=====")
	// log 的输出带时间戳，写 stderr，适合记录运行信息
	log.SetFlags(0) // 学习时关掉时间戳，输出更干净
	log.Println("这是一条日志（默认写到 stderr）")

	// 命令行程序的错误处理范式：
	//   出错 -> 打印到 stderr -> 以非 0 退出码结束
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("\n正常结束，退出码 0")
}

// run 把真正的业务逻辑收在一个可返回错误的函数里，
// 这样 main 只负责"调用 + 处理退出码"，也让逻辑可被测试。
func run() error {
	if len(os.Args) > 1 && os.Args[1] == "fail" {
		return errors.New("你故意让我失败的（试试 make run p=01-basics/07-modules-and-imports fail）")
	}
	return nil
}
