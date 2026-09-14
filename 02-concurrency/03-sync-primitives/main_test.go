package main

import (
	"fmt"
	"slices"
	"sync"
	"sync/atomic"
	"testing"
)

// ---------------------------------------------------------------------------
// Mutex：不变量一致性
// ---------------------------------------------------------------------------

func TestAccountConcurrentWithdraw(t *testing.T) {
	acct := NewAccount(1000)

	var wg sync.WaitGroup
	for range 100 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			acct.Withdraw(10)
		}()
	}
	wg.Wait()

	balance, ledger, rejected := acct.Snapshot()

	if balance != 0 {
		t.Errorf("余额 = %d, 想要 0", balance)
	}
	if len(ledger) != 100 {
		t.Errorf("流水 = %d 条, 想要 100 条", len(ledger))
	}
	if rejected != 0 {
		t.Errorf("不该有被拒的取款，实际 %d", rejected)
	}
	// 核心断言：余额和流水必须一致 —— 这就是"不变量"
	if balance != 1000-int64(len(ledger))*10 {
		t.Error("余额与流水不一致，说明锁没覆盖完整的不变量")
	}
}

func TestAccountOverdraft(t *testing.T) {
	acct := NewAccount(100)

	var wg sync.WaitGroup
	success := atomic.Int64{}
	// 200 个 goroutine 各取 10 元，但只有 100 元 -> 只能成功 10 次
	for range 200 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if acct.Withdraw(10) {
				success.Add(1)
			}
		}()
	}
	wg.Wait()

	if got := success.Load(); got != 10 {
		t.Errorf("成功取款 %d 次, 想要 10 次（超额取款说明有竞态）", got)
	}
	balance, _, rejected := acct.Snapshot()
	if balance != 0 {
		t.Errorf("余额 = %d, 想要 0", balance)
	}
	if rejected != 190 {
		t.Errorf("被拒 %d 次, 想要 190", rejected)
	}
}

func TestAccountDepositWithdraw(t *testing.T) {
	acct := NewAccount(0)

	var wg sync.WaitGroup
	for range 50 {
		wg.Add(2)
		go func() { defer wg.Done(); acct.Deposit(100) }()
		go func() { defer wg.Done(); acct.Withdraw(50) }()
	}
	wg.Wait()

	balance, ledger, _ := acct.Snapshot()
	// 50 次存 100 = 5000；取款次数取决于时序，但余额必须等于 5000 - 50×成功次数
	if balance < 0 {
		t.Errorf("余额不该为负: %d", balance)
	}
	if balance%50 != 0 {
		t.Errorf("余额 %d 不是 50 的倍数，说明有部分更新丢失", balance)
	}
	if len(ledger) == 0 {
		t.Error("应该有流水")
	}
}

// Snapshot 返回的必须是拷贝，否则调用方读取时会和内部 append 产生竞态。
func TestSnapshotReturnsCopy(t *testing.T) {
	acct := NewAccount(100)
	acct.Deposit(50)

	_, ledger1, _ := acct.Snapshot()
	acct.Deposit(25) // 内部会 append，可能复用底层数组
	_, ledger2, _ := acct.Snapshot()

	if len(ledger1) != 1 {
		t.Errorf("第一次快照应该有 1 条，实际 %d（返回了共享切片）", len(ledger1))
	}
	if len(ledger2) != 2 {
		t.Errorf("第二次快照应该有 2 条，实际 %d", len(ledger2))
	}
}

// ---------------------------------------------------------------------------
// RWMutex
// ---------------------------------------------------------------------------

func TestConfigCacheConcurrent(t *testing.T) {
	cache := NewConfigCache()

	var wg sync.WaitGroup
	// 并发写
	for i := range 50 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cache.Set(fmt.Sprintf("k%d", i), fmt.Sprintf("v%d", i))
		}()
	}
	// 并发读（可能读到也可能读不到，但绝不能崩或竞态）
	for range 50 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cache.Get("k1")
		}()
	}
	wg.Wait()

	if v, ok := cache.Get("k1"); !ok || v != "v1" {
		t.Errorf("Get(k1) = (%q, %t), 想要 (v1, true)", v, ok)
	}
	if _, ok := cache.Get("不存在"); ok {
		t.Error("不存在的 key 应该返回 false")
	}
}

// ---------------------------------------------------------------------------
// sync.Once
// ---------------------------------------------------------------------------

func TestSingletonOnce(t *testing.T) {
	// 注意：这个测试依赖包级状态，所以只能跑一次。
	// 真实项目里应该把 Once 封装进结构体，避免测试之间互相影响。
	const n = 100
	got := make([]*Singleton, n)

	var wg sync.WaitGroup
	for i := range n {
		wg.Add(1)
		go func() {
			defer wg.Done()
			got[i] = GetSingleton()
		}()
	}
	wg.Wait()

	// 所有 goroutine 必须拿到同一个指针
	for i := 1; i < n; i++ {
		if got[i] != got[0] {
			t.Fatalf("第 %d 个实例和第一个不同，Once 失效了", i)
		}
	}
	if c := InitCount(); c != 1 {
		t.Errorf("初始化执行了 %d 次, 想要 1 次", c)
	}
}

// 用结构体封装 Once，这样每个测试都能有独立实例
type lazyResource struct {
	once sync.Once
	val  string
	n    atomic.Int32
}

func (l *lazyResource) Get() string {
	l.once.Do(func() {
		l.n.Add(1)
		l.val = "初始化完成"
	})
	return l.val
}

func TestOncePerInstance(t *testing.T) {
	var l lazyResource

	var wg sync.WaitGroup
	for range 50 {
		wg.Add(1)
		go func() { defer wg.Done(); l.Get() }()
	}
	wg.Wait()

	if got := l.n.Load(); got != 1 {
		t.Errorf("初始化 %d 次, 想要 1 次", got)
	}
	if got := l.Get(); got != "初始化完成" {
		t.Errorf("Get() = %q", got)
	}
}

// ---------------------------------------------------------------------------
// atomic
// ---------------------------------------------------------------------------

func TestMetricsAtomic(t *testing.T) {
	var m Metrics

	var wg sync.WaitGroup
	const n = 1000
	for range n {
		wg.Add(1)
		go func() {
			defer wg.Done()
			m.Enter()
			m.RecordRequest(10)
			m.Leave()
		}()
	}
	wg.Wait()

	req, errs, bytes, active := m.Snapshot()
	if req != n {
		t.Errorf("请求数 = %d, 想要 %d（丢更新说明不是原子的）", req, n)
	}
	if bytes != n*10 {
		t.Errorf("字节数 = %d, 想要 %d", bytes, n*10)
	}
	if errs != 0 {
		t.Errorf("错误数 = %d, 想要 0", errs)
	}
	if active != 0 {
		t.Errorf("活跃数 = %d, 想要 0（Enter/Leave 应该配平）", active)
	}
}

// 对比：不加保护的计数器会丢更新。
//
// 这个演示**故意**制造数据竞争，所以不能放在测试里 ——
// 一旦用 -race 跑，整个测试套件都会因为检测到竞态而失败。
// 它被放在同目录的 unsafe_demo.go（带 //go:build ignore 标签，默认不参与构建）。
// 想亲眼看竞态报告就跑：
//   go run 02-concurrency/03-sync-primitives/unsafe_demo.go
//   go run -race 02-concurrency/03-sync-primitives/unsafe_demo.go

// ---------------------------------------------------------------------------
// sync.Map
// ---------------------------------------------------------------------------

func TestSessionStore(t *testing.T) {
	store := &SessionStore{}

	var wg sync.WaitGroup
	const n = 200
	for i := range n {
		wg.Add(1)
		go func() {
			defer wg.Done()
			store.Set(fmt.Sprintf("s%d", i), fmt.Sprintf("u%d", i))
		}()
	}
	wg.Wait()

	if got := store.Count(); got != n {
		t.Errorf("Count = %d, 想要 %d", got, n)
	}

	if u, ok := store.Get("s42"); !ok || u != "u42" {
		t.Errorf("Get(s42) = (%q, %t)", u, ok)
	}
	if _, ok := store.Get("不存在"); ok {
		t.Error("不存在的 key 应该返回 false")
	}

	store.Delete("s42")
	if _, ok := store.Get("s42"); ok {
		t.Error("删除后不该还能读到")
	}
	if got := store.Count(); got != n-1 {
		t.Errorf("删除后 Count = %d, 想要 %d", got, n-1)
	}
}

// ---------------------------------------------------------------------------
// channel 版计数器
// ---------------------------------------------------------------------------

func TestCounterChannel(t *testing.T) {
	c := NewCounterChannel()
	defer c.Close()

	var wg sync.WaitGroup
	const n = 500
	for range n {
		wg.Add(1)
		go func() { defer wg.Done(); c.Inc() }()
	}
	wg.Wait()

	if got := c.Value(); got != n {
		t.Errorf("Value = %d, 想要 %d", got, n)
	}
}

// ---------------------------------------------------------------------------
// 基准对比：Mutex vs atomic vs channel
// ---------------------------------------------------------------------------

func BenchmarkMutexCounter(b *testing.B) {
	var mu sync.Mutex
	var n int
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			mu.Lock()
			n++
			mu.Unlock()
		}
	})
	_ = n
}

func BenchmarkAtomicCounter(b *testing.B) {
	var n atomic.Int64
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			n.Add(1)
		}
	})
	_ = n.Load()
}

func BenchmarkChannelCounter(b *testing.B) {
	c := NewCounterChannel()
	defer c.Close()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			c.Inc()
		}
	})
}

func BenchmarkRWMutexRead(b *testing.B) {
	cache := NewConfigCache()
	cache.Set("k", "v")
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			cache.Get("k")
		}
	})
}

func BenchmarkMutexRead(b *testing.B) {
	var mu sync.Mutex
	m := map[string]string{"k": "v"}
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			mu.Lock()
			_ = m["k"]
			mu.Unlock()
		}
	})
}

// 确保 slices 被用到（保留导入）
var _ = slices.Equal[[]int]
