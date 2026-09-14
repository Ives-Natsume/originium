// Package greeter 演示"一个目录 = 一个包"以及导出规则。
//
// 包注释写在 package 声明正上方（中间不能有空行），
// 这样 `go doc` 和 IDE 悬浮提示里都能看到。
package greeter

import (
	"fmt"
	"strings"
)

// Prefix 是**导出**标识符（首字母大写），包外可以访问：greeter.Prefix
const Prefix = "你好"

// DefaultGreeting 也是导出的。
var DefaultGreeting = Prefix + ", 世界!"

// version 是**未导出**的（首字母小写），只有本包内部能用。
// 别的地方写 greeter.version 会直接编译报错。
const version = "1.0.0"

// init 函数：包被导入时自动执行，用来做初始化。
//
// 几个要点：
//   1. 可以有多个 init，按文件名的字典序依次执行
//   2. 无法被显式调用，也不能带参数/返回值
//   3. 导入顺序是"依赖优先"：被依赖的包先初始化完，再初始化依赖它的
//
// 实践建议：init 里只做"轻量的、确定不会失败"的事。
// 一旦 init 里 panic，整个程序起不来，而且很难排查（栈里看不到业务代码）。
// 需要传参或可能失败的初始化，用显式的 NewXxx() 函数。
//
// 为什么没用它？下面这个 log 变量演示了 init 的典型用途，但我们先注释掉，
// 免得每个用到这个包的地方都打印一行。
// func init() { log.Println("greeter 包已初始化") }

// Greet 返回一句问候。name 为空时使用默认值。
func Greet(name string) string {
	if strings.TrimSpace(name) == "" {
		return DefaultGreeting
	}
	return Prefix + ", " + strings.TrimSpace(name) + "!"
}

// Shout 返回大写版问候。因为它是导出的，调用方可以用 greeter.Shout(x)。
func Shout(name string) string {
	// 调用同包内的未导出函数，包外看不见这个函数
	return strings.ToUpper(adorn(Greet(name)))
}

// Version 把未导出的 version 暴露出去。
//
// 这是很常见的模式：内部实现细节私有，对外只给一个受控的读取入口。
func Version() string { return version }

// adorn 未导出：包外完全看不到，所以以后想改就改，不算破坏兼容性。
func adorn(s string) string {
	return "*** " + s + " ***"
}

// Stats 返回一些包级别的信息，用来演示跨包调用链。
func Stats() string {
	return fmt.Sprintf("greeter v%s, 默认问候语 = %q", version, DefaultGreeting)
}
