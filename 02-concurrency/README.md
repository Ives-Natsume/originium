# 02 · 并发

Go 最出名的特性。但并发不是"用了就快"的魔法，用好它需要建立几个关键直觉。

```bash
make run p=02-concurrency/01-goroutine-basics
make test-race p=02-concurrency/01-goroutine-basics   # 并发代码必须加 -race 跑
```

> **重点**：这个章节的每次测试都请用 `make test-race`（等价于 `go test -race`）。
> 竞态检测器只有在数据竞争**真的发生**时才会报告，所以测试要跑到才行。

| 模块 | 主题 | 关键知识点 |
| --- | --- | --- |
| `01-goroutine-basics` | goroutine 与 WaitGroup | 主 goroutine 退出即结束程序、`WaitGroup` 三件套、循环变量捕获、`GOMAXPROCS` |
| `02-channels` | channel | 无缓冲/有缓冲、`close` 语义、`range`、`select`、方向类型、谁发送谁关闭 |
| `03-sync-primitives` | 同步原语 | `Mutex`/`RWMutex`、`Once`、`atomic`、什么时候不该用锁 |
| `04-worker-pool` | 并发模式 | worker pool、fan-out/fan-in、pipeline、超时保护 |
| `05-context` | context | 取消传播、`WithTimeout`、`WithValue`、规范用法与常见误用 |
| `06-race-and-deadlock` | 竞态与死锁 | 如何读竞态报告、死锁的几种形态、排查思路 |

## 这一章要建立的直觉

1. **不要通过共享内存来通信，而要通过通信来共享内存。**
   —— 优先用 channel 传递所有权，而不是给共享变量加锁。
2. **goroutine 很便宜，但不是免费的**。每个至少 2KB 栈 + 调度开销，且会泄漏。
3. **goroutine 泄漏是最常见的生产事故**：启动了却永远不退出。
   每个 goroutine 都要能回答"它在什么条件下会结束？"
4. **能不用并发就别用**。串行代码好读、好调、好维护。只有在真的需要并行或需要异步等待时才上。
5. **锁保护的是"不变量"，不是"变量"**。加锁粒度、加锁范围，想清楚再写。

## 三句话判断你有没有真的懂

- 无缓冲 channel 的"发送"和"接收"，哪一方会阻塞？都阻塞会怎样？
- 为什么"关闭 channel"表示"不会再有数据了"，而不是"我读完了"？
- `for i := 0; i < 3; i++ { go func() { fmt.Println(i) }() }` 在 Go 1.22 前后输出不同，为什么？

## 学习顺序建议

先 `01` 和 `02` 打基础 → `03` 了解锁 → `04` 看真实模式 →
`05` 学 context（这是服务端开发的必修课）→ `06` 学会看竞态报告。

`06` 里有两个**故意写错**的示例，跑起来会看到真实的 panic / 竞态报告。
看完记得把 `-race` 加到你的日常测试命令里。