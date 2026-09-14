package main

import (
	"slices"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestWaitAndPrint(t *testing.T) {
	names := []string{"a", "b", "c", "d"}
	got := WaitAndPrint(names)

	if len(got) != len(names) {
		t.Fatalf("期望 %d 条结果，实际 %d: %v", len(names), len(got), got)
	}
	slices.Sort(got)
	want := []string{"处理完: a", "处理完: b", "处理完: c", "处理完: d"}
	if !slices.Equal(got, want) {
		t.Errorf("got %v, 想要 %v", got, want)
	}
}

// 验证确实是"并发"而不是"串行"：4 个 10ms 任务应该明显快于 40ms。
func TestWaitAndPrintIsConcurrent(t *testing.T) {
	start := time.Now()
	WaitAndPrint([]string{"a", "b", "c", "d"})
	elapsed := time.Since(start)

	if elapsed > 35*time.Millisecond {
		t.Errorf("耗时 %v，看起来是串行执行的（4 × 10ms = 40ms）", elapsed)
	}
}

func TestLoopCapture(t *testing.T) {
	// 如果循环变量被共享（Go 1.22 之前的行为），
	// 这里会出现重复值、缺失值。
	got := LoopCapture(100)
	want := make([]int, 100)
	for i := range want {
		want[i] = i
	}
	if !slices.Equal(got, want) {
		t.Errorf("循环变量捕获有问题: 长度 %d，前几个 %v", len(got), got[:min(10, len(got))])
	}
}

func TestLoopCaptureEmpty(t *testing.T) {
	if got := LoopCapture(0); len(got) != 0 {
		t.Errorf("n=0 应该返回空，实际 %v", got)
	}
}

func TestParallelSum(t *testing.T) {
	nums := make([]int, 10000)
	for i := range nums {
		nums[i] = i
	}
	var want int
	for _, n := range nums {
		want += n
	}

	// 不同 worker 数必须得到相同结果 —— 这是并行代码最基本的正确性要求
	for _, w := range []int{1, 2, 3, 4, 7, 16, 100} {
		if got := ParallelSum(nums, w); got != want {
			t.Errorf("workers=%d: 结果 %d, 想要 %d", w, got, want)
		}
	}
}

func TestParallelSumEdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		nums    []int
		workers int
		want    int
	}{
		{"空切片", nil, 4, 0},
		{"单元素", []int{7}, 4, 7},
		{"workers 大于元素数", []int{1, 2, 3}, 100, 6},
		{"workers 为 0 用默认值", []int{1, 2, 3}, 0, 6},
		{"workers 为负用默认值", []int{1, 2, 3}, -5, 6},
		{"含负数", []int{-1, -2, 3}, 2, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParallelSum(tt.nums, tt.workers); got != tt.want {
				t.Errorf("ParallelSum(%v, %d) = %d, 想要 %d", tt.nums, tt.workers, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// 对照实验：加锁 vs 原子操作 vs 无保护（错误）
// 用 `make bench p=02-concurrency/01-goroutine-basics` 跑出数据
// ---------------------------------------------------------------------------

// CounterMutex 用互斥锁保护计数器。
type CounterMutex struct {
	mu sync.Mutex
	n  int
}

func (c *CounterMutex) Inc() { c.mu.Lock(); c.n++; c.mu.Unlock() }
func (c *CounterMutex) Load() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.n
}

// CounterAtomic 用原子操作。单变量自增时比锁更快。
type CounterAtomic struct {
	n atomic.Int64
}

func (c *CounterAtomic) Inc()      { c.n.Add(1) }
func (c *CounterAtomic) Load() int { return int(c.n.Load()) }

const benchIters = 1000

func BenchmarkCounterMutex(b *testing.B) {
	for b.Loop() {
		var c CounterMutex
		var wg sync.WaitGroup
		for range benchIters {
			wg.Add(1)
			go func() { defer wg.Done(); c.Inc() }()
		}
		wg.Wait()
		if c.Load() != benchIters {
			b.Fatalf("计数错误: %d", c.Load())
		}
	}
}

func BenchmarkCounterAtomic(b *testing.B) {
	for b.Loop() {
		var c CounterAtomic
		var wg sync.WaitGroup
		for range benchIters {
			wg.Add(1)
			go func() { defer wg.Done(); c.Inc() }()
		}
		wg.Wait()
		if c.Load() != benchIters {
			b.Fatalf("计数错误: %d", c.Load())
		}
	}
}

// 测试本身也是并发的：用 t.Parallel() 让多个子测试并行跑。
func TestCounterImplementations(t *testing.T) {
	tests := []struct {
		name string
		c    interface {
			Inc()
			Load() int
		}
	}{
		{"Mutex", &CounterMutex{}},
		{"Atomic", &CounterAtomic{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel() // 与其他子测试并行执行

			var wg sync.WaitGroup
			for range benchIters {
				wg.Add(1)
				go func() { defer wg.Done(); tt.c.Inc() }()
			}
			wg.Wait()

			if got := tt.c.Load(); got != benchIters {
				t.Errorf("%s 并发自增丢更新: %d != %d（这就是数据竞争）", tt.name, got, benchIters)
			}
		})
	}
}
