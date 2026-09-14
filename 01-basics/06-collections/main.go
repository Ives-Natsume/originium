// Package main —— 第 6 课：切片、map 与泛型。
//
// 运行: make run p=01-basics/06-collections
package main

import (
	"cmp"
	"fmt"
	"maps"
	"slices"
	"strings"
)

// ============================================================================
// 一、slice 的内存模型（这一节是 Go 最容易出 bug 的地方）
// ============================================================================
//
// 一个 slice 变量本质上是三个字的"描述符"：{指向底层数组的指针, len, cap}
//
//	slice → [ptr | len | cap] → [ ][ ][ ][ ][ ]  底层数组
//
// 拷贝 slice 变量 = 拷贝描述符（三个字），底层数组是**共享**的。
// 所以函数内修改元素会影响外面，但函数内 append 导致扩容后会"断开连接"。

// AppendAndSee 演示共享底层数组带来的副作用。
func AppendAndSee(s []int) []int {
	s[0] = 999          // 改元素：影响调用方（同一底层数组）
	return append(s, 1) // append：容量够就复用数组（也影响调用方），不够就新开数组
}

// SafeAppend 演示正确做法：不修改入参，返回新切片。
// 库函数应该这样做 —— 调用方不该为"我的切片被偷改了"头疼。
func SafeAppend(s []int, vs ...int) []int {
	out := make([]int, 0, len(s)+len(vs))
	out = append(out, s...)
	out = append(out, vs...)
	return out
}

// RemoveAt 删除下标 i 的元素（保持顺序，O(n)）。
func RemoveAt(s []int, i int) []int {
	if i < 0 || i >= len(s) {
		return s
	}
	// 把 i 后面的所有元素左移一格，然后缩短长度
	copy(s[i:], s[i+1:])
	return s[:len(s)-1]
}

// RemoveAtSwap 删除下标 i（不保顺序，O(1)）。适合"集合"语义的切片。
func RemoveAtSwap(s []int, i int) []int {
	if i < 0 || i >= len(s) {
		return s
	}
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}

// Compact 原地去掉相邻重复（对应 slices.Compact）。
func Compact(s []int) []int {
	if len(s) == 0 {
		return s
	}
	w := 1
	for r := 1; r < len(s); r++ {
		if s[r] != s[w-1] {
			s[w] = s[r]
			w++
		}
	}
	return s[:w]
}

// ============================================================================
// 二、map
// ============================================================================

// WordCount 统计词频。map 的经典用法。
func WordCount(text string) map[string]int {
	counts := make(map[string]int)           // 别用 var m map[string]int —— 那是 nil，写入会 panic！
	for _, w := range strings.Fields(text) { // Fields 按任意空白切分，自动跳过连续空白
		w = strings.Trim(w, ".,!?;:\"'()[]") // 简单去掉常见标点
		if w == "" {
			continue
		}
		counts[strings.ToLower(w)]++
	}
	return counts
}

// GroupBy 按 key 归组，演示 map[string][]T。
func GroupBy[T any, K comparable](items []T, key func(T) K) map[K][]T {
	out := make(map[K][]T, len(items))
	for _, it := range items {
		k := key(it)
		out[k] = append(out[k], it)
	}
	return out
}

// Set 用 map[T]struct{} 表示集合。
// 为什么不是 map[T]bool？因为 struct{}{} 占 0 字节，省内存，且语义更明确。
type Set[T comparable] map[T]struct{}

func NewSet[T comparable](items ...T) Set[T] {
	s := make(Set[T], len(items))
	for _, it := range items {
		s.Add(it)
	}
	return s
}

func (s Set[T]) Add(v T)      { s[v] = struct{}{} }
func (s Set[T]) Has(v T) bool { _, ok := s[v]; return ok }
func (s Set[T]) Len() int     { return len(s) }
func (s Set[T]) Delete(v T)   { delete(s, v) }

// Items 返回所有元素，顺序不确定。
//
// 这里刻意不排序：Set[T comparable] 只要求 T 可比较，不一定可排序
// （比如 T 是 struct 或指针就没法排）。
// 想排序用下面的 SortedItems —— 这就是"只在需要时收紧约束"：
// 约束写在函数上，而不是整个类型上。
func (s Set[T]) Items() []T {
	return slices.Collect(maps.Keys(s))
}

// SortedItems 返回排好序的元素，要求 T 可排序。
func SortedItems[T cmp.Ordered](s Set[T]) []T {
	return slices.Sorted(maps.Keys(s))
}

func (s Set[T]) Union(o Set[T]) Set[T] {
	out := NewSet[T]()
	for k := range s {
		out.Add(k)
	}
	for k := range o {
		out.Add(k)
	}
	return out
}
func (s Set[T]) Intersect(o Set[T]) Set[T] {
	out := NewSet[T]()
	for k := range s {
		if o.Has(k) {
			out.Add(k)
		}
	}
	return out
}

// ============================================================================
// 三、泛型
// ============================================================================

// Number 是类型约束：~ 表示"底层类型是这些之一的都行"。
// 没有 ~ 的话，type MyInt int 这种自定义类型就不满足约束了。
type Number interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~float32 | ~float64
}

// SumAll 泛型求和：一份代码支持所有数值类型。
func SumAll[T Number](nums []T) T {
	var total T
	for _, n := range nums {
		total += n
	}
	return total
}

// Map 泛型转换：[]T -> []U。
func Map[T, U any](s []T, f func(T) U) []U {
	out := make([]U, 0, len(s))
	for _, v := range s {
		out = append(out, f(v))
	}
	return out
}

// Filter 泛型筛选。
func Filter[T any](s []T, keep func(T) bool) []T {
	out := make([]T, 0, len(s))
	for _, v := range s {
		if keep(v) {
			out = append(out, v)
		}
	}
	return out
}

// Reduce 泛型归约。
func Reduce[T, U any](s []T, init U, f func(acc U, v T) U) U {
	acc := init
	for _, v := range s {
		acc = f(acc, v)
	}
	return acc
}

// Stack 泛型数据结构：一个类型参数就够通用。
type Stack[T any] struct {
	items []T
}

func (s *Stack[T]) Push(v T) { s.items = append(s.items, v) }

func (s *Stack[T]) Pop() (T, bool) {
	var zero T
	if len(s.items) == 0 {
		return zero, false // 用 ok 而不是抛异常，符合 Go 风格
	}
	v := s.items[len(s.items)-1]
	s.items = s.items[:len(s.items)-1]
	return v, true
}

func (s *Stack[T]) Peek() (T, bool) {
	var zero T
	if len(s.items) == 0 {
		return zero, false
	}
	return s.items[len(s.items)-1], true
}

func (s *Stack[T]) Len() int { return len(s.items) }

// Max 泛型 + 约束：T 必须可比较大小。
func Max[T Number](a, b T) T {
	if a > b {
		return a
	}
	return b
}

func main() {
	fmt.Println("===== 1. slice 是'描述符'，底层数组共享 =====")

	s := []int{1, 2, 3, 4, 5}
	fmt.Printf("s      = %v  len=%d cap=%d\n", s, len(s), cap(s))

	// 切片操作：a 和 s 共享同一块底层数组
	a := s[1:3]
	fmt.Printf("a=s[1:3] = %v  len=%d cap=%d  ← cap 会一直延伸到 s 的末尾\n", a, len(a), cap(a))

	// 改 a 的元素，s 也跟着变 —— 同一个底层数组
	a[0] = 20
	fmt.Printf("改完 a[0] 后 s = %v   ← 被「顺带」改了\n", s)

	// append：容量够就复用数组，会污染 s
	s = []int{1, 2, 3, 4, 5}
	a = s[1:3]
	a = append(a, 99)
	fmt.Printf("append 到 a 后 s = %v  ← s[3] 被覆盖成 99 了！\n", s)

	// 三下标切片：显式限制容量，append 就一定会新开数组
	s = []int{1, 2, 3, 4, 5}
	b := s[1:3:3] // [low:high:max]，cap = max-low = 2
	fmt.Printf("b=s[1:3:3] len=%d cap=%d\n", len(b), cap(b))
	b = append(b, 99)
	fmt.Printf("append 到 b 后 s = %v  ← s 没被污染，因为扩容了新数组\n", s)

	fmt.Println("\n结论：函数收到 slice 时，改元素会影响调用方。")
	fmt.Println("     不想被影响就自己 make + copy（见 SafeAppend）。")

	fmt.Println("\n===== 2. len 与 cap 的增长 =====")
	var g []int
	prevCap := cap(g)
	for i := range 20 {
		g = append(g, i)
		if cap(g) != prevCap {
			fmt.Printf("  len=%2d 时 cap 从 %2d 涨到 %2d\n", len(g), prevCap, cap(g))
			prevCap = cap(g)
		}
	}
	fmt.Println("  结论：append 的扩容是指数式的，所以反复 append 的均摊成本是 O(1)")
	fmt.Println("  但频繁扩容会有内存拷贝 —— 已知大小就用 make([]T, 0, n) 预分配")

	fmt.Println("\n===== 3. 修改切片元素的三种方式 =====")
	nums := []int{1, 2, 3, 4, 5, 6}

	fmt.Printf("原切片          = %v\n", nums)
	n2 := RemoveAt(slices.Clone(nums), 2)
	fmt.Printf("RemoveAt(2)     = %v  ← 保序，O(n)\n", n2)
	n3 := RemoveAtSwap(slices.Clone(nums), 2)
	fmt.Printf("RemoveAtSwap(2) = %v  ← 不保序，O(1)\n", n3)

	dup := []int{1, 1, 2, 2, 2, 3, 1}
	fmt.Printf("Compact(%v) = %v  ← 只去相邻重复\n", dup, Compact(slices.Clone(dup)))

	fmt.Println("\n===== 4. map 基础 =====")
	text := "go is great, go is fast. go go go"
	counts := WordCount(text)

	// 遍历顺序是**随机的** —— 这是故意的设计，防止你依赖顺序
	fmt.Println("词频（注意每次运行顺序都不一样）:")
	for w, n := range counts {
		fmt.Printf("  %-6s %d\n", w, n)
	}
	// 需要顺序输出就自己排
	fmt.Println("按词排序:")
	for _, w := range slices.Sorted(maps.Keys(counts)) {
		fmt.Printf("  %-6s %d\n", w, counts[w])
	}

	// comma-ok 惯用法
	if n, ok := counts["go"]; ok {
		fmt.Printf("\ncomma-ok: \"go\" 出现 %d 次\n", n)
	}
	if _, ok := counts["python"]; !ok {
		fmt.Println("comma-ok: \"python\" 不存在（ok=false，而不是 0）")
	}
	// 直接读不存在的 key 得到零值，不报错 —— 所以"0"和"不存在"必须靠 ok 区分
	fmt.Printf("直接读 counts[\"python\"] = %d  ← 零值，无从判断是否存在\n", counts["python"])

	fmt.Println("\n===== 5. map 作为集合 =====")
	aSet := NewSet(1, 2, 3, 4)
	bSet := NewSet(3, 4, 5, 6)
	fmt.Printf("A           = %v\n", SortedItems(aSet))
	fmt.Printf("B           = %v\n", SortedItems(bSet))
	fmt.Printf("A ∪ B       = %v\n", SortedItems(aSet.Union(bSet)))
	fmt.Printf("A ∩ B       = %v\n", SortedItems(aSet.Intersect(bSet)))
	fmt.Printf("A 有 3 吗?   %t\n", aSet.Has(3))
	fmt.Printf("Items() 不保证顺序: %v\n", aSet.Items())

	fmt.Println("\n===== 6. GroupBy =====")
	type Emp struct {
		Name string
		Dept string
	}
	emps := []Emp{
		{"Neo", "研发"}, {"李雷", "研发"}, {"韩梅梅", "产品"},
		{"Alex", "研发"}, {"Bob", "产品"},
	}
	groups := GroupBy(emps, func(e Emp) string { return e.Dept })
	for _, dept := range slices.Sorted(maps.Keys(groups)) {
		names := Map(groups[dept], func(e Emp) string { return e.Name })
		fmt.Printf("  %-4s: %v\n", dept, names)
	}

	fmt.Println("\n===== 7. 泛型 =====")
	ints := []int{1, 2, 3, 4, 5}
	floats := []float64{1.5, 2.5, 3.0}
	fmt.Printf("SumAll(ints)   = %d\n", SumAll(ints))
	fmt.Printf("SumAll(floats) = %v\n", SumAll(floats))
	fmt.Printf("Max(3, 9)      = %d\n", Max(3, 9))
	fmt.Printf("Max(1.5, 0.5)  = %v\n", Max(1.5, 0.5))

	// Map / Filter / Reduce 像乐高一样组合
	doubled := Map(ints, func(n int) int { return n * 2 })
	evens := Filter(doubled, func(n int) bool { return n%4 == 0 })
	total := Reduce(evens, 0, func(acc, n int) int { return acc + n })
	fmt.Printf("ints --×2--> %v --筛4的倍数--> %v --求和--> %d\n", doubled, evens, total)

	// 泛型也能直接和标准库组合
	strs := Map(ints, func(n int) string { return fmt.Sprintf("第%d", n) })
	fmt.Printf("Map 成字符串: %v\n", strs)
	fmt.Printf("用标准库排序: %v\n", slices.Sorted(slices.Values(strs)))

	fmt.Println("\n===== 8. 泛型数据结构 =====")
	var st Stack[string]
	st.Push("第一")
	st.Push("第二")
	st.Push("第三")
	fmt.Printf("stack len=%d\n", st.Len())
	if top, ok := st.Peek(); ok {
		fmt.Printf("peek = %q（不弹出）\n", top)
	}
	for {
		v, ok := st.Pop()
		if !ok {
			break
		}
		fmt.Printf("  pop -> %q\n", v)
	}
	if _, ok := st.Pop(); !ok {
		fmt.Println("空栈 Pop 返回 ok=false，不会 panic")
	}

	fmt.Println("\n===== 9. 标准库 slices / maps 速查 =====")
	d := []int{3, 1, 4, 1, 5, 9, 2, 6}
	fmt.Printf("slices.Max(%v)      = %d\n", d, slices.Max(d))
	fmt.Printf("slices.Min          = %d\n", slices.Min(d))
	fmt.Printf("slices.Contains 4?  = %t\n", slices.Contains(d, 4))
	fmt.Printf("slices.Index(9)     = %d\n", slices.Index(d, 9))
	fmt.Printf("slices.Clone        = %v（独立副本）\n", slices.Clone(d))
	fmt.Printf("排序前              = %v\n", d)
	slices.Sort(d)
	fmt.Printf("slices.Sort 后      = %v\n", d)
	fmt.Printf("binary search 5     = %d（下标，-1 表示没有）\n", slices.Index(d, 5))
	fmt.Printf("去重后              = %v\n", slices.Compact(slices.Clone(d)))
}
