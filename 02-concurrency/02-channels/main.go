// Package main —— 并发第 2 课：channel。
//
// 运行:      make run p=02-concurrency/02-channels
// 竞态检测:  make test-race p=02-concurrency/02-channels
package main

import (
	"fmt"
	"slices"
	"sync"
	"time"
)

// ============================================================================
// 一、channel 的两个基本类型
// ============================================================================
//
//	ch := make(chan int)     // 无缓冲：发送方会阻塞，直到有接收方接手
//	ch := make(chan int, 3)  // 有缓冲：缓冲没满就不阻塞
//
// 无缓冲 channel 是"同步点"：它保证发送和接收在时间上对齐（happens-before）。
// 有缓冲 channel 是"队列"：解耦了生产者和消费者的节奏。

// PingPong 演示无缓冲 channel 的"交接"语义。
//
// 每一轮：主 goroutine 发 -> worker 收 -> worker 发 -> 主 goroutine 收。
// 因为是无缓冲的，两边会严格交替，输出顺序是确定的。
func PingPong(rounds int) []string {
	ping := make(chan int) // 主 -> worker
	pong := make(chan int) // worker -> 主
	var log []string

	go func() {
		for range rounds {
			n := <-ping // 阻塞等主 goroutine 发；收到后对方才解除阻塞
			pong <- n * n
		}
	}()

	for i := range rounds {
		ping <- i
		log = append(log, fmt.Sprintf("ping %d -> pong %d", i, <-pong))
	}
	return log
}

// BufferedDemo 演示有缓冲 channel：发送方可以在没人接收时先"垫"进缓冲区。
func BufferedDemo(capacity int) []int {
	ch := make(chan int, capacity)
	for i := range capacity {
		ch <- i // 缓冲区没满 -> 不阻塞
	}
	// 此时 channel 里有 capacity 个元素，但没有任何接收者

	var out []int
	close(ch) // 关闭后仍能读出剩余数据
	for v := range ch {
		out = append(out, v)
	}
	return out
}

// ============================================================================
// 二、close 的语义（这是最容易搞错的地方）
// ============================================================================
//
// 一句话记牢：**close 表示"不会再有新值了"，而不是"我读完了"。**
//
// 规则：
//   1. 只有发送方应该 close，接收方永远不要 close
//   2. 向已关闭的 channel 发送 -> panic
//   3. 从已关闭的 channel 接收 -> 立刻返回零值，ok=false
//   4. 重复 close -> panic；close(nil) -> panic
//   5. 有多个发送方时，用单独的"协调 channel"或 WaitGroup 决定何时 close

// CloseSemantics 展示关闭后的读取行为。
func CloseSemantics() (values []int, closedReads int, open bool) {
	ch := make(chan int, 3)
	ch <- 1
	ch <- 2
	close(ch)

	for v := range ch { // range 会在 channel 关闭且读空后自动结束
		values = append(values, v)
	}

	// 关闭后再读：立刻返回零值 0，且 ok=false
	v, ok := <-ch
	if !ok {
		closedReads = v // 约定用这个变量名装"关闭后读到的值"
	}
	return values, closedReads, ok
}

// ============================================================================
// 三、select：多路复用
// ============================================================================
//
// select 会等待"任意一个 case 就绪"。如果多个同时就绪，随机挑一个（故意如此，
// 防止你依赖固定顺序）。加 default 就变成非阻塞。

// Fastest 从多个 channel 里取最先就绪的那个。
// 用"先到先得"实现竞速：谁先返回就用谁的结果。
func Fastest(urls []string) (string, time.Duration) {
	type result struct {
		url string
		d   time.Duration
	}
	ch := make(chan result, len(urls))

	for i, u := range urls {
		go func() {
			// 模拟不同的响应延迟
			d := time.Duration(20+i*30) * time.Millisecond
			time.Sleep(d)
			ch <- result{url: u, d: d}
		}()
	}

	// 只取第一个结果，后面的 goroutine 会在 buffered channel 上写入后结束
	// （因为有缓冲，它们不会永远阻塞 —— 这是防 goroutine 泄漏的常用技巧）
	r := <-ch
	return r.url, r.d
}

// NonBlockingRecv 演示 select + default 的"能拿就拿，不拿就走"。
func NonBlockingRecv(ch <-chan int) (int, bool) {
	select {
	case v := <-ch:
		return v, true
	default: // 没有就绪的 case 时立刻走这里，绝不阻塞
		return 0, false
	}
}

// NonBlockingSend 演示"能发就发，发不出就丢弃"。
// 常见场景：日志/监控这类"丢了也不影响主流程"的数据。
func NonBlockingSend(ch chan<- int, v int) bool {
	select {
	case ch <- v:
		return true
	default:
		return false
	}
}

// Timeout 演示 select + time.After 做超时控制。
// 注意：生产代码更推荐用 context（见 05-context），
// 因为 time.After 每次调用都会创建一个定时器，高频调用会积累垃圾。
func Timeout(work func() string, limit time.Duration) (string, error) {
	done := make(chan string, 1) // 缓冲 1：即使超时了，worker 也能把结果塞进去而不泄漏
	go func() { done <- work() }()

	select {
	case r := <-done:
		return r, nil
	case <-time.After(limit):
		return "", fmt.Errorf("超时（%v）", limit)
	}
}

// ============================================================================
// 四、方向类型：用类型来约束"谁能读写"
// ============================================================================
//
//	chan T    可读可写
//	chan<- T  只能发送（send-only）
//	<-chan T  只能接收（receive-only）
//
// 函数签名里写明方向是很好的习惯：
//   * 自我说明（这个函数只生产 / 只消费）
//   * 编译器帮你检查，写错了直接编译失败

// producer 只发送。
func producer(out chan<- int, nums []int) {
	for _, n := range nums {
		out <- n
	}
	close(out) // 发送方负责 close
}

// square 接收和发送，做转换。
func square(in <-chan int, out chan<- int) {
	for v := range in { // 上游 close 后这个循环自动结束
		out <- v * v
	}
	close(out)
}

// ============================================================================
// 五、fan-in / fan-out（并发模式的基石）
// ============================================================================

// FanIn 把多个 channel 合并成一个。
//
// 泛型版本：T 可以是任何类型，所以 int / string / 自定义结构体都能用。
// 这里用 WaitGroup + 关闭信号，是标准的 fan-in 写法。
func FanIn[T any](chans ...<-chan T) <-chan T {
	out := make(chan T)
	var wg sync.WaitGroup

	for _, ch := range chans {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for v := range ch {
				out <- v
			}
		}()
	}

	// 关键：等所有输入都耗尽后，才能关闭输出。
	// 谁创建 out，谁负责关闭 —— 用一个 goroutine 做这件事。
	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}

// Generate 把切片转成 channel（生成器模式）。
func Generate(nums ...int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for _, n := range nums {
			out <- n
		}
	}()
	return out
}

// ToSlice 把 channel 收集成切片（消费端）。
func ToSlice(ch <-chan int) []int {
	var out []int
	for v := range ch {
		out = append(out, v)
	}
	return out
}

func main() {
	fmt.Println("===== 1. 无缓冲 channel：同步交接 =====")
	for _, line := range PingPong(3) {
		fmt.Printf("  %s\n", line)
	}
	fmt.Println("  无缓冲 = 发送方一直阻塞到有人接收，所以两边严格交替")
	fmt.Println("  它是**同步点**：这也是它最有价值的语义，不只是'慢'")

	fmt.Println("\n===== 2. 有缓冲 channel：解耦节奏 =====")
	fmt.Printf("  缓冲区填 5 个再用 range 读出: %v\n", BufferedDemo(5))
	fmt.Println("  有缓冲 = 允许生产者跑到消费者前面，代价是数据'在途'")
	fmt.Println("  经验：容量应该反映你想容忍多少积压，别随便写个 100")

	fmt.Println("\n===== 3. close 的语义 =====")
	values, closedVal, ok := CloseSemantics()
	fmt.Printf("  range 读出全部: %v\n", values)
	fmt.Printf("  关闭后再读: 值=%d ok=%t  ← 立刻返回，不阻塞\n", closedVal, ok)
	fmt.Println("  三条铁律:")
	fmt.Println("    · 只有发送方 close，接收方绝不 close")
	fmt.Println("    · close 后不能发（panic），但可以继续收（读到零值）")
	fmt.Println("    · 重复 close / close(nil) / 向 nil channel 发送 → panic")

	fmt.Println("\n===== 4. select 多路复用 =====")
	url, d := Fastest([]string{"server-a", "server-b", "server-c"})
	fmt.Printf("  竞速结果: %s（用了 %v）← 最快的那个\n", url, d)

	ch := make(chan int, 1)
	emptyV, emptyOK := NonBlockingRecv(ch)
	fmt.Printf("  空 channel 非阻塞读: v=%d ok=%t\n", emptyV, emptyOK)
	ch <- 42
	v, ok := NonBlockingRecv(ch)
	fmt.Printf("  有数据时读: v=%d ok=%t\n", v, ok)

	dropped := 0
	for i := range 10 {
		// 缓冲区容量 1，所以只有第一个能进，其余被丢弃
		if !NonBlockingSend(ch, i) {
			dropped++
		}
	}
	fmt.Printf("  非阻塞发送丢弃了 %d 个（缓冲区已满时直接放弃）\n", dropped)

	fmt.Println("\n===== 5. 超时控制 =====")
	fast := func() string { time.Sleep(10 * time.Millisecond); return "很快的结果" }
	slow := func() string { time.Sleep(200 * time.Millisecond); return "慢结果" }

	if r, err := Timeout(fast, 100*time.Millisecond); err == nil {
		fmt.Printf("  快任务: %s\n", r)
	}
	if _, err := Timeout(slow, 50*time.Millisecond); err != nil {
		fmt.Printf("  慢任务: %v  ← 调用方没有被无限拖住\n", err)
	}
	fmt.Println("  注意 done 用的是带缓冲 channel：超时后 worker 仍能写进去，不会泄漏 goroutine")

	fmt.Println("\n===== 6. 方向类型 =====")
	fmt.Println("  func producer(out chan<- int)  只能发送")
	fmt.Println("  func consumer(in <-chan int)   只能接收")
	fmt.Println("  写反了编译不过 —— 免费的正确性保障")

	fmt.Println("\n===== 7. pipeline：把数据流串起来 =====")
	fmt.Println("  数据流向: 生成 -> 平方 -> 筛选 -> 收集")

	// 每一段都是"上一个的输出 = 下一个的输入"
	gen := Generate(1, 2, 3, 4, 5, 6, 7, 8)
	sqd := make(chan int)
	go square(gen, sqd)

	// 中间插一段筛选
	filtered := make(chan int)
	go func() {
		defer close(filtered)
		for v := range sqd {
			if v%2 == 0 { // 只留偶数
				filtered <- v
			}
		}
	}()

	got := ToSlice(filtered)
	fmt.Printf("  1..8 平方后取偶数: %v\n", got)
	fmt.Println("  pipeline 的好处：每段职责单一、可单独测试、天然并发")
	fmt.Println("  坏处：goroutine 多了之后，'谁负责 close'容易乱 —— 记住'谁创建谁关闭'")

	fmt.Println("\n===== 8. fan-in：多路合并 =====")
	c1 := Generate(1, 3, 5)
	c2 := Generate(2, 4, 6)
	c3 := Generate(7, 8, 9)

	merged := ToSlice(FanIn(c1, c2, c3))
	slices.Sort(merged)
	fmt.Printf("  三个 channel 合并后: %v\n", merged)
	fmt.Println("  输出顺序不确定（取决于调度），需要顺序就自己排序")

	fmt.Println("\n===== 9. 一句话总结用哪个 =====")
	fmt.Println("  需要同步握手/信号  -> 无缓冲 channel")
	fmt.Println("  需要解耦生产消费    -> 有缓冲 channel")
	fmt.Println("  需要等待一组任务    -> sync.WaitGroup（比 channel 更直接）")
	fmt.Println("  需要保护共享状态    -> sync.Mutex / atomic（见第 3 课）")
	fmt.Println("  需要取消/超时传播   -> context（见第 5 课）")
}
