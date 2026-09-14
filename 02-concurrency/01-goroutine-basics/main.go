// Package main —— 并发第 1 课：goroutine 与 WaitGroup。
//
// 运行:      make run p=02-concurrency/01-goroutine-basics
// 竞态检测:  make test-race p=02-concurrency/01-goroutine-basics
package main

import (
	"fmt"
	"runtime"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

// ============================================================================
// 一、goroutine 的基本行为
// ============================================================================
//
// `go f(x)` 表示"在新 goroutine 里执行 f(x)"。
//
// 三个必须记住的事实：
//   1. main 函数返回时，所有还在跑的 goroutine 会被**立刻杀掉**（不会等它们）
//   2. goroutine 不阻塞启动方，所以"启动了"不等于"做完了"
//   3. 循环变量捕获问题（Go 1.22 起已修复，看下面 LoopCapture）

// WaitAndPrint 演示正确的等待：用 WaitGroup 等所有 goroutine 结束。
//
// WaitGroup 的用法固定三步，记住它就不会错：
//
//	wg.Add(n)  —— 声明"我要等 n 个任务"，必须在启动 goroutine **之前**调用
//	wg.Done()  —— 每个任务结束时调用（通常 defer wg.Done()）
//	wg.Wait()  —— 阻塞直到计数器归零
func WaitAndPrint(names []string) []string {
	var (
		wg      sync.WaitGroup
		mu      sync.Mutex // 保护 results：多个 goroutine 会并发写入
		results []string
	)

	for _, name := range names {
		wg.Add(1)
		go func() {
			defer wg.Done() // defer 保证即使 panic 也会减计数
			// 模拟一点工作
			time.Sleep(10 * time.Millisecond)

			mu.Lock() // 临界区：同一时刻只允许一个 goroutine 进入
			results = append(results, "处理完: "+name)
			mu.Unlock()
		}()
	}

	wg.Wait() // 等所有 goroutine 结束，之后 results 不再被并发访问
	return results
}

// FireAndForget 是**反面教材**：启动了 goroutine 却不等它。
// 主函数一返回，这个 goroutine 就被杀掉了，输出可能根本不出现。
// 取消注释看看效果 —— 这就是"goroutine 泄漏/丢失"的雏形。
func FireAndForget() {
	go func() {
		time.Sleep(50 * time.Millisecond)
		fmt.Println("  [反面教材] 我大概永远不会被打印出来")
	}()
	// 没有 Wait：main 结束 -> 进程退出 -> 这个 goroutine 直接消失
}

// LoopCapture 演示循环变量捕获。
//
// Go 1.22 之前：循环变量 i 是**共享的同一个**，闭包捕获的是它本身，
//
//	所以三个 goroutine 可能都打印 3。
//
// Go 1.22 及以后：每次迭代都会创建新的 i，所以每个 goroutine 拿到自己的值。
//
// 本仓库 go.mod 声明 go 1.27，所以是"新"行为。
// 但如果你看老代码/老项目，很可能要求你写 go func(i int){...}(i) 这种显式传参。
func LoopCapture(n int) []int {
	var (
		wg  sync.WaitGroup
		mu  sync.Mutex
		out []int
	)
	for i := range n {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mu.Lock()
			out = append(out, i) // Go 1.22+ 每个 i 独立
			mu.Unlock()
		}()
	}
	wg.Wait()
	sort.Ints(out)
	return out
}

// ============================================================================
// 二、并发 vs 并行，以及 GOMAXPROCS
// ============================================================================
//
// 并发(concurrency) = 同时处理多个任务（结构化，不一定是同一时刻）
// 并行(parallelism) = 同一时刻真的在多核上一起跑
//
// 并发是"写法"，并行是"运行效果"。Go 的 M:N 调度器把 G(goroutine) 映射到
// M(系统线程) 上，由 P(处理器) 提供运行资源，P 的数量 = GOMAXPROCS。

// ParallelSum 把切片切段并行求和。
//
// 注意：这里的注释会告诉你**什么时候并行没意义** ——
// 这个函数对小切片比串行还慢，因为启动 goroutine 和合并结果都有开销。
// 并行的收益 = 单次计算时间 × 任务数 - 调度开销。
// 计算太简单时，收益为负。真正的性能优化永远先 profile（见 make bench）。
func ParallelSum(nums []int, workers int) int {
	if len(nums) == 0 {
		return 0
	}
	if workers <= 0 {
		workers = runtime.GOMAXPROCS(0)
	}

	chunk := (len(nums) + workers - 1) / workers
	var (
		wg    sync.WaitGroup
		total atomic.Int64 // 原子操作：无锁的并发安全累加
	)

	for start := 0; start < len(nums); start += chunk {
		end := min(start+chunk, len(nums))
		wg.Add(1)
		go func(part []int) {
			defer wg.Done()
			var local int64
			for _, n := range part {
				local += int64(n)
			}
			total.Add(local) // 各自算局部和，最后原子加一次，减少争用
		}(nums[start:end])
	}

	wg.Wait()
	return int(total.Load())
}

func main() {
	fmt.Printf("GOMAXPROCS = %d（逻辑 CPU 数 %d）\n", runtime.GOMAXPROCS(0), runtime.NumCPU())
	fmt.Printf("当前 goroutine 数 = %d\n", runtime.NumGoroutine())

	// ================= 1. 主 goroutine 退出 = 程序结束 =================
	fmt.Println("\n===== 1. 主 goroutine 退出会杀掉所有 goroutine =====")
	FireAndForget()
	fmt.Println("（反面教材已启动，但 main 马上要干别的活，它可能来不及打印）")

	// ================= 2. WaitGroup 正确等待 =================
	fmt.Println("\n===== 2. 用 WaitGroup 正确等待 =====")
	start := time.Now()
	results := WaitAndPrint([]string{"Neo", "李雷", "Alex", "Bob"})
	elapsed := time.Since(start)
	sort.Strings(results)
	for _, r := range results {
		fmt.Printf("  %s\n", r)
	}
	fmt.Printf("4 个任务各耗时 10ms，总耗时 %v ← 是并发的，不是 40ms\n", elapsed.Round(time.Millisecond))

	// ================= 3. 循环变量捕获 =================
	fmt.Println("\n===== 3. 循环变量捕获（Go 1.22+ 每个迭代独立）=====")
	fmt.Printf("LoopCapture(5) = %v  ← 应该完整包含 0..4，没有重复\n", LoopCapture(5))

	// ================= 4. 并发数控制 =================
	fmt.Println("\n===== 4. goroutine 很便宜，但不是免费的 =====")
	before := runtime.NumGoroutine()
	var wg sync.WaitGroup
	const n = 10000
	for range n {
		wg.Add(1)
		go func() { defer wg.Done() }() // 空任务，只为测量创建开销
	}
	mid := runtime.NumGoroutine()
	wg.Wait()
	fmt.Printf("启动前 %d 个 → 启动 %d 个空 goroutine 时约 %d 个 → 结束后 %d 个\n",
		before, n, mid, runtime.NumGoroutine())
	fmt.Println("结论：初始栈只有 2KB 左右，V8/Linux 线程是 MB 级，所以 Go 能开几十万个")
	fmt.Println("      但每个都要调度、都要栈，无节制开 goroutine 一样会 OOM / 拖慢调度")

	// ================= 5. 并行求和 =================
	fmt.Println("\n===== 5. 并行求和（以及它什么时候不该用）=====")
	nums := make([]int, 1_000_000)
	for i := range nums {
		nums[i] = i % 100
	}
	serial := func() int {
		s := 0
		for _, n := range nums {
			s += n
		}
		return s
	}

	t0 := time.Now()
	var want int
	want = serial()
	serialDur := time.Since(t0)

	parResults := make(map[int]int)
	parDur := make(map[int]time.Duration)
	for _, w := range []int{1, 2, 4, 8} {
		t0 := time.Now()
		got := ParallelSum(nums, w)
		parDur[w] = time.Since(t0)
		parResults[w] = got
		if got != want {
			fmt.Printf("  ✗ workers=%d 结果错了: %d != %d\n", w, got, want)
		}
	}
	fmt.Printf("串行            %v  (sum=%d)\n", serialDur.Round(time.Microsecond), want)
	for _, w := range []int{1, 2, 4, 8} {
		fmt.Printf("并行 workers=%d  %v  (sum=%d)\n", w, parDur[w].Round(time.Microsecond), parResults[w])
	}
	fmt.Println("注意：小数据量时并行反而更慢（调度开销 > 计算收益）。")
	fmt.Println("      性能问题要用 make bench / pprof 量出来再优化，别凭感觉猜。")
	fmt.Println("      make bench p=02-concurrency/01-goroutine-basics 有对照基准测试。")
}
