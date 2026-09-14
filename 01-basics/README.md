# 01 · Go 基础

语言本身的地基：语法、类型系统、函数、错误处理、结构体与接口。

每个模块都是一个**可以独立运行**的小程序，并且配了测试。建议的学习方式：

```bash
make run p=01-basics/01-hello     # 先跑一遍看输出
# 然后打开代码读，改一改，再跑
make test p=01-basics/01-hello    # 测试你的理解对不对
make watch p=01-basics/01-hello   # 边改边自动重跑
```

| 模块 | 主题 | 关键知识点 |
| --- | --- | --- |
| `01-hello` | 第一行 Go | `package`/`main`、`os.Args`、导出规则（大小写）、退出码 |
| `02-fmt-and-io` | 输入输出 | `fmt` 动词与对齐、`Sprintf`、`io.Reader`/`io.Writer`、`bufio` |
| `03-types-and-control` | 类型与控制流 | 数值类型与溢出、`rune`/`byte`、`const`+`iota`、`switch`、带标签 `break` |
| `04-func-and-error` | 函数与错误 | 多返回值、变参、闭包、`defer`、错误包装 `%w`、`errors.Is/As`、`panic/recover` |
| `05-struct-interface` | 结构体与接口 | 方法集、嵌入、接口隐式实现、类型断言、`sort`、`Stringer` |
| `06-collections` | 集合与泛型 | slice 内存模型与陷阱、map、泛型函数、`slices`/`maps` 标准库 |
| `07-modules-and-imports` | 模块与包 | `go.mod`、导入路径、`internal/` 可见性规则、包的拆分粒度 |

## 这一章要建立的直觉

1. **零值可用**。Go 的变量默认值（`0`、`""`、`nil`、`false`）是精心设计过的，很多时候不需要构造函数。
2. **错误是值**，不是异常。`if err != nil` 看着啰嗦，但控制流是显式的、可组合的。
3. **接口由使用方定义**，实现方不需要声明"我实现了某接口"。这是 Go 和其他语言最大的思路差异。
4. **值语义 vs 引用语义**：结构体赋值是拷贝，slice/map 是"描述符"（共享底层数组）。想清楚这一点，很多 bug 就消失了。

## 学完自测

- 能说清 `[]int` 传进函数后，函数内 `append` 会不会影响外面的 len？
- 能解释 `for i, r := range "中文"` 里 `i` 为什么是跳跃的？
- 知道什么时候该返回 `error`，什么时候该 `panic`？
- 能自己写一个 interface，让两个不同的 struct 都满足它？

答不上来就回到对应模块，`main_test.go` 里有针对性用例。