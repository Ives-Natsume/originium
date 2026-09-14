// Package main —— 并发第 3 课：同步原语。
//
// 运行:      make run p=02-concurrency/03-sync-primitives
// 竞态检测:  make test-race p=02-concurrency/03-sync-primitives
package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// ============================================================================
// 一、Mutex：保护"不变量"，不是保护"变量"
// ============================================================================
//
// 加锁的核心问题不是"哪个变量要保护"，而是"哪些操作必须原子地一起完成"。
//
// 反例：一个银行账户，余额和流水必须同时更新。
// 如果只锁余额、不锁流水，就会出现"钱扣了但流水没记"的中间状态被别的
// goroutine 观察到。锁要覆盖**整个不变量**。

// Account 用互斥锁保护余额与流水的一致性。
type Account struct {
	mu       sync.Mutex
	balance  int64
	ledger   []string // 流水
	rejected int64    // 被拒绝的取款次数
}

func NewAccount(initial int64) *Account {
	return &Account{balance: initial}
}

// Withdraw 取款。返回是否成功。
//
// 注意锁的范围：从"检查余额"到"扣款 + 记流水"必须是一个整体，
// 否则两个 goroutine 可能同时通过余额检查，导致超额取款。
func (a *Account) Withdraw(amount int64) bool {
	a.mu.Lock()
	defer a.mu.Unlock() // defer 解锁：即使中间 panic 也不会死锁

	if amount <= 0 || a.balance < amount {
		a.rejected++
		return false
	}
	a.balance -= amount
	a.ledger = append(a.ledger, fmt.Sprintf("取款 %d，余额 %d", amount, a.balance))
	return true
}

// Deposit 存款。
func (a *Account) Deposit(amount int64) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if amount <= 0 {
		return
	}
	a.balance += amount
	a.ledger = append(a.ledger, fmt.Sprintf("存款 %d，余额 %d", amount, a.balance))
}

// Snapshot 一次性读取多个字段。
//
// 关键：返回的是**拷贝**，不是内部切片。
// 如果直接返回 a.ledger，调用方拿到的是共享底层数组，
// 之后我们 append 时就会和调用方的读取产生竞态。
func (a *Account) Snapshot() (balance int64, ledger []string, rejected int64) {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.balance, append([]string(nil), a.ledger...), a.rejected
}

// ============================================================================
// 二、RWMutex：读多写少时用
// ============================================================================
//
// 规则：
//   RLock  —— 读锁，多个 goroutine 可以同时持有
//   Lock   —— 写锁，独占，且会等待所有读锁释放
//
// 什么时候值得用？读操作远多于写操作（比如配置缓存、路由表）。
// 什么时候不值得？读写差不多，或者临界区极短 —— RWMutex 本身开销比 Mutex 大。

// ConfigCache 是一个读多写少的缓存。
type ConfigCache struct {
	mu   sync.RWMutex
	data map[string]string
	hits atomic.Int64
}

func NewConfigCache() *ConfigCache {
	return &ConfigCache{data: make(map[string]string)}
}

// Get 用读锁：多个 goroutine 可以并发读。
func (c *ConfigCache) Get(key string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	v, ok := c.data[key]
	if ok {
		c.hits.Add(1)
	}
	return v, ok
}

// Set 用写锁：独占。
func (c *ConfigCache) Set(key, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[key] = value
}

// Hits 返回命中次数。
func (c *ConfigCache) Hits() int64 { return c.hits.Load() }

// ============================================================================
// 三、sync.Once：只执行一次
// ============================================================================
//
// 典型用途：单例、懒加载、只初始化一次的资源。
// 即使多个 goroutine 同时调用，函数也只会执行一次，其余调用会阻塞到它完成。

// Singleton 演示线程安全的懒加载。
type Singleton struct {
	ID string
}

var (
	singletonOnce sync.Once
	singleton     *Singleton
	singletonInit int32 // 记录 init 实际执行了几次（用于测试）
)

// GetSingleton 无论多少 goroutine 并发调用，初始化只发生一次。
func GetSingleton() *Singleton {
	singletonOnce.Do(func() {
		atomic.AddInt32(&singletonInit, 1)
		time.Sleep(5 * time.Millisecond) // 模拟耗时初始化
		singleton = &Singleton{ID: "唯一实例"}
	})
	return singleton
}

// InitCount 返回初始化实际执行次数（测试用）。
func InitCount() int32 { return atomic.LoadInt32(&singletonInit) }

// ============================================================================
// 四、atomic：单变量的无锁操作
// ============================================================================
//
// 适用：计数器、标志位、指针替换这类"单个值"的读写。
// 不适用：多个字段需要保持一致（那必须用锁）。
//
// 为什么快？因为它用 CPU 的原子指令，不需要操作系统介入、不会让 goroutine 睡眠。

// Metrics 用原子操作实现无锁指标统计。
type Metrics struct {
	requests atomic.Int64
	errors   atomic.Int64
	bytes    atomic.Int64
	active   atomic.Int64
}

func (m *Metrics) RecordRequest(n int64) {
	m.requests.Add(1)
	m.bytes.Add(n)
}

func (m *Metrics) RecordError() { m.errors.Add(1) }

func (m *Metrics) Enter() { m.active.Add(1) }
func (m *Metrics) Leave() { m.active.Add(-1) }

// Snapshot 读取多个原子变量。
//
// 注意：这不是"一致性快照"！读取过程中别的 goroutine 可能已经改了某些值，
// 所以 requests 和 errors 可能来自不同时刻。
// 如果业务要求"这几个数必须严格对应同一时刻"，那就得用锁。
func (m *Metrics) Snapshot() (requests, errors, bytes, active int64) {
	return m.requests.Load(), m.errors.Load(), m.bytes.Load(), m.active.Load()
}

// ============================================================================
// 五、sync.Map：只在特定场景才用
// ============================================================================
//
// 官方文档说得很清楚：大多数场景应该用 map + Mutex。
// sync.Map 适合这两种情况：
//   1. key 集合基本固定，主要是读（比如缓存）
//   2. 多个 goroutine 读写**不相交**的 key 集合
//
// 它的代价：接口类型装箱、没有 len() 的原子保证、遍历是弱一致的。

// SessionStore 用 sync.Map 存会话。
type SessionStore struct {
	m sync.Map
}

func (s *SessionStore) Set(id, user string) { s.m.Store(id, user) }

func (s *SessionStore) Get(id string) (string, bool) {
	v, ok := s.m.Load(id)
	if !ok {
		return "", false
	}
	return v.(string), true // sync.Map 存的是 any，取出来要断言
}

func (s *SessionStore) Delete(id string) { s.m.Delete(id) }

// Count 遍历统计。注意：遍历期间其他 goroutine 的修改可能看到也可能看不到。
func (s *SessionStore) Count() int {
	n := 0
	s.m.Range(func(_, _ any) bool {
		n++
		return true // 返回 false 可以提前结束遍历
	})
	return n
}

// ============================================================================
// 六、什么时候**不该**用锁
// ============================================================================

// CounterChannel 用 channel 实现计数器 —— 演示"用通信代替共享内存"。
//
// 对比 CounterMutex：这个版本把状态**独占**在一个 goroutine 里，
// 外部只能通过 channel 请求它修改。好处是永远不会忘记加锁，
// 坏处是每次操作都要走 channel，比 Mutex 慢。
//
// 结论：简单计数器用 atomic，复杂状态用 Mutex，
//
//	"状态机 + 事件流"这种天然串行的逻辑才适合 channel。
type CounterChannel struct {
	inc  chan struct{}
	get  chan int
	stop chan struct{}
}

func NewCounterChannel() *CounterChannel {
	c := &CounterChannel{
		inc:  make(chan struct{}),
		get:  make(chan int),
		stop: make(chan struct{}),
	}
	go c.run() // 状态只被这一个 goroutine 碰
	return c
}

func (c *CounterChannel) run() {
	n := 0
	for {
		select {
		case <-c.inc:
			n++
		case c.get <- n: // 把当前值发出去
		case <-c.stop:
			return
		}
	}
}

func (c *CounterChannel) Inc()       { c.inc <- struct{}{} }
func (c *CounterChannel) Value() int { return <-c.get }
func (c *CounterChannel) Close()     { close(c.stop) }

func main() {
	fmt.Println("===== 1. Mutex 保护不变量 =====")
	acct := NewAccount(1000)

	var wg sync.WaitGroup
	// 100 个 goroutine 各取 10 元：总共想取 1000，余额刚好够
	for range 100 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			acct.Withdraw(10)
		}()
	}
	wg.Wait()

	balance, ledger, rejected := acct.Snapshot()
	fmt.Printf("初始 1000，100 次取 10 元后: 余额=%d 流水条数=%d 被拒=%d\n",
		balance, len(ledger), rejected)
	fmt.Println("关键：余额和流水条数必须一致（1000-100×10=0，流水 100 条）")
	fmt.Println("      如果锁的范围没覆盖'检查+扣款+记账'，就会出现超额取款")

	fmt.Println("\n===== 2. RWMutex：读多写少 =====")
	cache := NewConfigCache()
	cache.Set("host", "localhost")
	cache.Set("port", "8080")

	var rwg sync.WaitGroup
	for range 100 { // 100 个并发读
		rwg.Add(1)
		go func() {
			defer rwg.Done()
			cache.Get("host")
		}()
	}
	rwg.Wait()
	fmt.Printf("100 次并发读，命中 %d 次\n", cache.Hits())
	fmt.Println("读锁之间不互斥，所以这些读是真正并行的")

	fmt.Println("\n===== 3. sync.Once：初始化只做一次 =====")
	var owg sync.WaitGroup
	ids := make([]string, 50)
	for i := range 50 {
		owg.Add(1)
		go func() {
			defer owg.Done()
			ids[i] = GetSingleton().ID // 50 个 goroutine 同时抢
		}()
	}
	owg.Wait()
	fmt.Printf("50 个 goroutine 拿到同一个实例: %q\n", ids[0])
	fmt.Printf("初始化函数实际执行了 %d 次（必须是 1）\n", InitCount())

	fmt.Println("\n===== 4. atomic：单变量无锁 =====")
	var m Metrics
	var mwg sync.WaitGroup
	for range 1000 {
		mwg.Add(1)
		go func() {
			defer mwg.Done()
			m.Enter()
			m.RecordRequest(100)
			if time.Now().UnixNano()%7 == 0 {
				m.RecordError()
			}
			m.Leave()
		}()
	}
	mwg.Wait()
	req, errs, bytes, active := m.Snapshot()
	fmt.Printf("请求=%d 错误=%d 字节=%d 活跃=%d\n", req, errs, bytes, active)
	fmt.Println("注意：Snapshot 不是一致性快照，各字段可能来自不同时刻")

	fmt.Println("\n===== 5. sync.Map =====")
	store := &SessionStore{}
	var swg sync.WaitGroup
	for i := range 100 {
		swg.Add(1)
		go func() {
			defer swg.Done()
			store.Set(fmt.Sprintf("sess-%d", i), fmt.Sprintf("user-%d", i))
		}()
	}
	swg.Wait()
	fmt.Printf("写入 100 个会话，Count() = %d\n", store.Count())
	if u, ok := store.Get("sess-42"); ok {
		fmt.Printf("读取 sess-42 -> %q\n", u)
	}
	store.Delete("sess-42")
	fmt.Printf("删除后 Count() = %d\n", store.Count())
	fmt.Println("提醒：大多数场景 map + Mutex 更好，sync.Map 只在特定场景有优势")

	fmt.Println("\n===== 6. 用 channel 代替锁 =====")
	cc := NewCounterChannel()
	defer cc.Close()

	var cwg sync.WaitGroup
	for range 100 {
		cwg.Add(1)
		go func() {
			defer cwg.Done()
			cc.Inc()
		}()
	}
	cwg.Wait()
	fmt.Printf("channel 版计数器 = %d\n", cc.Value())
	fmt.Println("状态被独占在一个 goroutine 里，永远不会忘记加锁")

	fmt.Println("\n===== 选型速查 =====")
	fmt.Println("  单个计数器/标志位        -> atomic")
	fmt.Println("  多个字段要保持一致        -> sync.Mutex")
	fmt.Println("  读远多于写                -> sync.RWMutex")
	fmt.Println("  只初始化一次              -> sync.Once")
	fmt.Println("  key 固定、以读为主        -> sync.Map")
	fmt.Println("  状态机/事件流             -> channel + 单个 owner goroutine")
}
