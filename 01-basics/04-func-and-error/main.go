// Package main —— 第 4 课：函数与错误处理。
//
// 运行: make run p=01-basics/04-func-and-error
package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// ============================================================================
// 一、多返回值：Go 没有异常，错误是"第二个返回值"
// ============================================================================

// 哨兵错误（sentinel error）：包级别导出的、可被调用方用 errors.Is 比较的错误值。
// 命名惯例是 Err 开头。
var (
	ErrDivByZero = errors.New("除数不能为零")
	ErrTooShort  = errors.New("太短了")
)

// Div 整除。第二个返回值是 error：nil 表示成功。
//
// 这是 Go 最核心的约定。调用方必须显式处理：
//
//	r, err := Div(10, 0)
//	if err != nil { ... }
//
// 看起来啰嗦，但控制流是看得见的 —— 不会像异常那样从看不见的地方跳出来。
func Div(a, b int) (int, error) {
	if b == 0 {
		return 0, ErrDivByZero
	}
	return a / b, nil
}

// SafeDiv 演示错误包装。
//
// %w 会保留原始错误，形成一条"错误链"：
//
//	计算 10/0 失败: 除数不能为零
//	                         ↑ 这一层还能被 errors.Is 找到
//
// 用 %v 就会丢掉链路，只留文字。
func SafeDiv(a, b int) (int, error) {
	r, err := Div(a, b)
	if err != nil {
		return 0, fmt.Errorf("计算 %d/%d 失败: %w", a, b, err)
	}
	return r, nil
}

// ============================================================================
// 二、自定义错误类型：需要携带结构化信息时用
// ============================================================================

// ValidationError 带上了"是哪个字段、什么值、因为什么"。
// 调用方可以用 errors.As 把它取出来，读 Field 做针对性处理。
type ValidationError struct {
	Field string
	Value string
	Err   error // 底层原因
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("字段 %s 取值 %q 不合法: %v", e.Field, e.Value, e.Err)
}

// Unwrap 让 errors.Is / errors.As 能继续往下钻。
// 有了这个方法，*ValidationError 就自动具备了错误链能力。
func (e *ValidationError) Unwrap() error { return e.Err }

// ValidateName 校验用户名。
func ValidateName(name string) error {
	if strings.TrimSpace(name) == "" {
		return &ValidationError{Field: "name", Value: name, Err: errors.New("不能为空")}
	}
	if n := len([]rune(name)); n < 2 {
		return &ValidationError{Field: "name", Value: name, Err: ErrTooShort}
	}
	return nil
}

// ============================================================================
// 三、变参函数与闭包
// ============================================================================

// Sum 变参：nums 在函数内是 []int。
func Sum(nums ...int) int {
	total := 0
	for _, n := range nums {
		total += n
	}
	return total
}

// Stats 演示"变参 + 多返回值"组合。
func Stats(nums ...int) (min, max, sum int, ok bool) {
	if len(nums) == 0 {
		return 0, 0, 0, false
	}
	min, max = nums[0], nums[0]
	for _, n := range nums {
		min = builtinMin(min, n)
		max = builtinMax(max, n)
		sum += n
	}
	return min, max, sum, true
}

func builtinMin(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func builtinMax(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Counter 返回一个闭包。
// 每次调用返回的函数，都会读到并修改同一个 count 变量 —— 闭包"捕获"了它。
// 注意：这意味着它不是并发安全的，多 goroutine 同时调用要加锁（见 02-concurrency）。
func Counter() func() int {
	count := 0
	return func() int {
		count++
		return count
	}
}

// Accumulator 返回一个"累加器"，演示闭包携带状态。
func Accumulator(initial int) (add func(int) int, reset func()) {
	total := initial
	add = func(delta int) int {
		total += delta
		return total
	}
	reset = func() { total = initial }
	return add, reset
}

// ============================================================================
// 四、defer：延迟执行，后进先出
// ============================================================================

// DeferOrder 演示 defer 的执行顺序：LIFO（栈）。
//
// 注意这里必须用**具名返回值** out：defer 是在 return 语句求值之后才执行的，
// 如果写成 `func DeferOrder() []string { var out []string; ...; return out }`，
// 返回的切片头在 return 那一刻就被拷贝走了，defer 里的 append 改的是局部变量，
// 调用方拿到的还是 nil。这是 defer 最经典的坑之一。
func DeferOrder() (out []string) {
	for i := 1; i <= 3; i++ {
		// 注意：i 是每次迭代的新变量（Go 1.22+），所以闭包捕获的值各不相同。
		defer func() { out = append(out, strconv.Itoa(i)) }()
	}
	return // 返回 [3 2 1]
}

// DeferArgs 演示"defer 的实参在 defer 那一刻就求值了"。
func DeferArgs() (log []string) {
	x := 1
	defer func(v int) {
		log = append(log, fmt.Sprintf("捕获时 v=%d", v))
	}(x) // x 此刻是 1，后面再改不影响
	x = 100
	log = append(log, fmt.Sprintf("函数结束时 x=%d", x))
	return // [捕获时 v=1, 函数结束时 x=100]
}

// NamedReturn 演示 defer 可以修改具名返回值。
// 这是 Go 独有的能力，常见于"出错时改返回值"。
func NamedReturn() (n int) {
	defer func() { n *= 2 }() // 在 return 之后、返回给调用方之前执行
	n = 5
	return // 实际返回 10
}

// WithCleanup 是 defer 最典型的用途：打开资源后立刻安排关闭。
// 这样即使中间 return 或 panic，资源也会释放。
func WithCleanup(work func() (string, error)) (result string, err error) {
	fmt.Println("  [资源] 打开（假装是文件/数据库连接/锁）")
	defer fmt.Println("  [资源] 关闭 ← defer 保证一定执行")
	return work()
}

// ============================================================================
// 五、panic / recover：只用于"真的不该发生"的情况
// ============================================================================

// MustParseInt 演示 panic 的合理用途：
// 在"配置写死在代码里、错了就是程序员的问题"这类场景，
// 让程序在启动时立刻崩溃，比带着坏值继续跑要好。
func MustParseInt(s string) int {
	n, err := strconv.Atoi(s)
	if err != nil {
		panic(fmt.Sprintf("MustParseInt(%q): %v", s, err))
	}
	return n
}

// SafeRun 把可能 panic 的函数包起来，转成 error。
//
// 什么时候**不该**用：业务逻辑里的"预期失败"（找不到记录、参数非法）
// 应该返回 error，而不是 panic。panic 只处理"程序有 bug"。
func SafeRun(fn func()) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("捕获到 panic: %v", r)
		}
	}()
	fn()
	return nil
}

func main() {
	fmt.Println("===== 1. 多返回值与错误包装 =====")
	if r, err := SafeDiv(10, 2); err == nil {
		fmt.Printf("10/2 = %d\n", r)
	}
	_, err := SafeDiv(10, 0)
	fmt.Printf("10/0 -> %v\n", err)

	// errors.Is：判断错误链里"有没有"某个哨兵错误
	fmt.Printf("errors.Is(err, ErrDivByZero) = %t\n", errors.Is(err, ErrDivByZero))
	fmt.Println("（这就是用 %w 的价值：信息加了上下文，但原始错误仍可识别）")

	fmt.Println("\n===== 2. 自定义错误类型 + errors.As =====")
	for _, name := range []string{"", "A", "Neo"} {
		err := ValidateName(name)
		if err == nil {
			fmt.Printf("name=%-4q ✓ 通过\n", name)
			continue
		}
		fmt.Printf("name=%-4q ✗ %v\n", name, err)

		// errors.As：把错误链里的 *ValidationError 取出来，读结构化字段
		var ve *ValidationError
		if errors.As(err, &ve) {
			fmt.Printf("         └─ 定位到字段 %q，底层原因: %v（是 ErrTooShort 吗? %t）\n",
				ve.Field, ve.Err, errors.Is(err, ErrTooShort))
		}
	}

	fmt.Println("\n===== 3. 变参 =====")
	fmt.Printf("Sum()        = %d\n", Sum())
	fmt.Printf("Sum(1,2,3)   = %d\n", Sum(1, 2, 3))
	nums := []int{1, 2, 3, 4}
	fmt.Printf("Sum(nums...) = %d   ← 用 ... 把切片展开\n", Sum(nums...))
	if mn, mx, sm, ok := Stats(nums...); ok {
		fmt.Printf("Stats -> min=%d max=%d sum=%d\n", mn, mx, sm)
	}

	fmt.Println("\n===== 4. 闭包 =====")
	next := Counter()
	fmt.Printf("Counter 三次调用: %d %d %d\n", next(), next(), next())
	// 每次调用 Counter() 都会创建**独立**的状态，互不干扰
	other := Counter()
	fmt.Printf("另一个独立的 Counter: %d\n", other())

	add, reset := Accumulator(100)
	add(10)
	add(20)
	fmt.Printf("累加器: %d\n", add(0))
	reset()
	fmt.Printf("reset 之后: %d\n", add(0))

	fmt.Println("\n===== 5. defer 的执行细节 =====")
	fmt.Printf("DeferOrder()  = %v   ← LIFO：最后 defer 的最先执行\n", DeferOrder())
	fmt.Printf("DeferArgs()   = %v   ← 实参在 defer 时求值\n", DeferArgs())
	fmt.Printf("NamedReturn() = %d    ← 具名返回值被 defer 改了\n", NamedReturn())

	fmt.Println("\n退出时的资源释放：")
	if s, err := WithCleanup(func() (string, error) { return "读完的内容", nil }); err == nil {
		fmt.Printf("  结果: %s\n", s)
	}

	fmt.Println("\n===== 6. panic / recover =====")
	fmt.Printf("MustParseInt(\"42\") = %d\n", MustParseInt("42"))

	err = SafeRun(func() { MustParseInt("不是数字") })
	fmt.Printf("SafeRun -> %v\n", err)

	// 演示一个真实会 panic 的操作：切片越界
	err = SafeRun(func() {
		s := []int{1, 2, 3}
		i := 10
		fmt.Println(s[i]) // 越界，panic
	})
	fmt.Printf("切片越界 -> %v\n", err)

	err = SafeRun(func() {
		var p *int
		fmt.Println(*p) // 解引用 nil 指针，panic
	})
	fmt.Printf("nil 解引用 -> %v\n", err)

	err = SafeRun(func() { panic("我故意的") })
	fmt.Printf("手动 panic -> %v\n", err)

	fmt.Println("\n记住原则：")
	fmt.Println("  预期内的失败 -> 返回 error")
	fmt.Println("  程序有 bug   -> panic（让它早点炸，别带着坏状态继续跑）")
	fmt.Println("  库的边界     -> 可以在 recover 后返回 error，别让 panic 穿透到调用方")
}
