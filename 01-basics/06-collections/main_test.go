package main

import (
	"fmt"
	"maps"
	"slices"
	"testing"
)

// ---------------------------------------------------------------------------
// slice 内存模型：这些测试把"共享底层数组"的行为固定下来，方便你对照理解
// ---------------------------------------------------------------------------

func TestSliceSharesBackingArray(t *testing.T) {
	t.Run("切片的元素修改会反映到原切片", func(t *testing.T) {
		s := []int{1, 2, 3, 4, 5}
		a := s[1:3]
		a[0] = 20
		if s[1] != 20 {
			t.Errorf("期望 s[1] == 20（共享底层数组），实际 %v", s)
		}
	})

	t.Run("容量够时 append 会污染原切片", func(t *testing.T) {
		s := []int{1, 2, 3, 4, 5}
		a := s[1:3]       // len=2 cap=4
		a = append(a, 99) // 容量够，直接写进 s[3]
		if s[3] != 99 {
			t.Errorf("容量够时 append 应该复用底层数组，s = %v", s)
		}
	})

	t.Run("三下标切片限制容量后 append 不会污染", func(t *testing.T) {
		s := []int{1, 2, 3, 4, 5}
		b := s[1:3:3]     // len=2 cap=2
		b = append(b, 99) // 容量不够 -> 新数组
		if s[3] != 4 {
			t.Errorf("限制容量后 append 不该影响 s，实际 s = %v", s)
		}
		if b[2] != 99 {
			t.Errorf("b 自己应该拿到新元素，实际 %v", b)
		}
	})

	t.Run("传参是描述符拷贝：append 出去的不影响调用方 len", func(t *testing.T) {
		s := make([]int, 3, 10)
		got := AppendAndSee(s)
		// 函数内 append 后 len 变成了 4，但调用方的 s 还是 3
		if len(s) != 3 {
			t.Errorf("调用方 len 不该变，实际 %d", len(s))
		}
		if len(got) != 4 {
			t.Errorf("返回值 len = %d, 想要 4", len(got))
		}
		// 但元素级别的修改是共享的
		if s[0] != 999 {
			t.Errorf("元素修改应该共享，实际 s[0] = %d", s[0])
		}
	})
}

func TestSafeAppend(t *testing.T) {
	orig := []int{1, 2, 3}
	got := SafeAppend(orig, 4, 5)

	if len(orig) != 3 {
		t.Errorf("原切片不能被改，实际 %v", orig)
	}
	want := []int{1, 2, 3, 4, 5}
	if !slices.Equal(got, want) {
		t.Errorf("SafeAppend = %v, 想要 %v", got, want)
	}
	// 确认底层数组确实是独立的
	got[0] = 100
	if orig[0] == 100 {
		t.Error("返回的切片应该有独立底层数组")
	}
}

func TestRemoveAt(t *testing.T) {
	tests := []struct {
		name string
		in   []int
		i    int
		want []int
	}{
		{"删中间", []int{1, 2, 3, 4}, 1, []int{1, 3, 4}},
		{"删头", []int{1, 2, 3}, 0, []int{2, 3}},
		{"删尾", []int{1, 2, 3}, 2, []int{1, 2}},
		{"越界返回原样", []int{1, 2}, 5, []int{1, 2}},
		{"负下标返回原样", []int{1, 2}, -1, []int{1, 2}},
		{"空切片", nil, 0, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RemoveAt(slices.Clone(tt.in), tt.i)
			if !slices.Equal(got, tt.want) {
				t.Errorf("RemoveAt(%v, %d) = %v, 想要 %v", tt.in, tt.i, got, tt.want)
			}
		})
	}
}

func TestRemoveAtSwap(t *testing.T) {
	got := RemoveAtSwap([]int{1, 2, 3, 4}, 1)
	// 不保序：把尾巴补到被删的位置 -> [1, 4, 3]
	want := []int{1, 4, 3}
	if !slices.Equal(got, want) {
		t.Errorf("RemoveAtSwap = %v, 想要 %v", got, want)
	}
	if len(got) != 3 {
		t.Errorf("长度应为 3，实际 %d", len(got))
	}
}

func TestCompact(t *testing.T) {
	tests := []struct {
		in   []int
		want []int
	}{
		{nil, nil},
		{[]int{}, []int{}},
		{[]int{1}, []int{1}},
		{[]int{1, 1, 2, 2, 2, 3}, []int{1, 2, 3}},
		{[]int{1, 2, 1}, []int{1, 2, 1}}, // 不相邻的不去
	}
	for _, tt := range tests {
		got := Compact(slices.Clone(tt.in))
		if !slices.Equal(got, tt.want) {
			t.Errorf("Compact(%v) = %v, 想要 %v", tt.in, got, tt.want)
		}
	}
}

// ---------------------------------------------------------------------------
// map
// ---------------------------------------------------------------------------

func TestWordCount(t *testing.T) {
	got := WordCount("go is great, go is fast. go go go")

	want := map[string]int{"go": 5, "is": 2, "great": 1, "fast": 1}
	if !maps.Equal(got, want) {
		t.Errorf("WordCount = %v, 想要 %v", got, want)
	}
}

func TestMapCommaOk(t *testing.T) {
	m := map[string]int{"a": 1}

	t.Run("存在的 key", func(t *testing.T) {
		v, ok := m["a"]
		if !ok || v != 1 {
			t.Errorf("got (%d, %t), 想要 (1, true)", v, ok)
		}
	})

	t.Run("不存在的 key：零值 + ok=false", func(t *testing.T) {
		v, ok := m["zzz"]
		if ok {
			t.Error("ok 应该是 false")
		}
		if v != 0 {
			t.Errorf("值应该是零值 0，实际 %d", v)
		}
	})

	t.Run("delete 之后 ok 变 false", func(t *testing.T) {
		delete(m, "a")
		if _, ok := m["a"]; ok {
			t.Error("delete 后不该还在")
		}
		// delete 不存在的 key 不报错，这是安全的
		delete(m, "never-existed")
	})
}

// nil map 只能读不能写 —— 这是新手很常见的 panic 来源
func TestNilMap(t *testing.T) {
	var m map[string]int

	// 读是安全的：nil map 的读返回零值
	if v, ok := m["x"]; ok || v != 0 {
		t.Errorf("nil map 读应该得到 (0, false)，实际 (%d, %t)", v, ok)
	}
	if got := len(m); got != 0 {
		t.Errorf("nil map len 应为 0，实际 %d", got)
	}

	// 但写会 panic
	defer func() {
		if r := recover(); r == nil {
			t.Error("向 nil map 写入应该 panic —— 所以要用 make(map[K]V)")
		}
	}()
	m["x"] = 1
}

func TestSet(t *testing.T) {
	s := NewSet(3, 1, 2)

	if s.Len() != 3 {
		t.Errorf("Len = %d, 想要 3", s.Len())
	}
	if !s.Has(1) || s.Has(9) {
		t.Error("Has 结果不对")
	}
	// SortedItems 保证有序，方便断言；Items() 的顺序是随机的
	if got := SortedItems(s); !slices.Equal(got, []int{1, 2, 3}) {
		t.Errorf("SortedItems = %v, 想要 [1 2 3]", got)
	}
	// Items() 的顺序不确定，所以只能比较集合内容（排序后再比）
	if got := s.Items(); !slices.Equal(slices.Sorted(slices.Values(got)), []int{1, 2, 3}) {
		t.Errorf("Items 内容不对: %v", got)
	}

	s.Delete(2)
	if s.Has(2) {
		t.Error("Delete 失败")
	}

	// 重复添加同一个元素不会变大
	s.Add(1)
	if s.Len() != 2 {
		t.Errorf("重复 Add 不该增加元素，实际 %d", s.Len())
	}

	a := NewSet(1, 2, 3)
	b := NewSet(3, 4)
	if got := SortedItems(a.Union(b)); !slices.Equal(got, []int{1, 2, 3, 4}) {
		t.Errorf("Union = %v", got)
	}
	if got := SortedItems(a.Intersect(b)); !slices.Equal(got, []int{3}) {
		t.Errorf("Intersect = %v", got)
	}
}

func TestGroupBy(t *testing.T) {
	got := GroupBy([]int{1, 2, 3, 4, 5, 6}, func(n int) string {
		if n%2 == 0 {
			return "偶"
		}
		return "奇"
	})

	if len(got["偶"]) != 3 || len(got["奇"]) != 3 {
		t.Errorf("分组数量不对: %v", got)
	}
	if !slices.Equal(got["奇"], []int{1, 3, 5}) {
		t.Errorf("奇数分组 = %v", got["奇"])
	}
}

// ---------------------------------------------------------------------------
// 泛型
// ---------------------------------------------------------------------------

func TestSumAll(t *testing.T) {
	// 同一份代码支持多种类型 —— 这就是泛型的价值
	if got := SumAll([]int{1, 2, 3}); got != 6 {
		t.Errorf("int 求和 = %d", got)
	}
	if got := SumAll([]float64{1.5, 2.5}); got != 4.0 {
		t.Errorf("float 求和 = %v", got)
	}
	if got := SumAll([]int64{1 << 40, 1 << 40}); got != 1<<41 {
		t.Errorf("int64 求和 = %d", got)
	}
	if got := SumAll([]int{}); got != 0 {
		t.Errorf("空切片应该返回零值，实际 %d", got)
	}
}

// 自定义类型只要底层类型满足约束（因为有 ~）就能用
type MyScore int

func TestSumAll_CustomType(t *testing.T) {
	got := SumAll([]MyScore{10, 20, 30})
	if got != MyScore(60) {
		t.Errorf("自定义类型求和 = %v, 想要 60", got)
	}
}

func TestMapFilterReduce(t *testing.T) {
	nums := []int{1, 2, 3, 4, 5, 6}

	doubled := Map(nums, func(n int) int { return n * 2 })
	if !slices.Equal(doubled, []int{2, 4, 6, 8, 10, 12}) {
		t.Errorf("Map = %v", doubled)
	}

	// Map 可以改变类型：int -> string
	strs := Map(nums, func(n int) string { return fmt.Sprintf("%d", n) })
	if !slices.Equal(strs, []string{"1", "2", "3", "4", "5", "6"}) {
		t.Errorf("Map 换类型 = %v", strs)
	}

	evens := Filter(nums, func(n int) bool { return n%2 == 0 })
	if !slices.Equal(evens, []int{2, 4, 6}) {
		t.Errorf("Filter = %v", evens)
	}

	sum := Reduce(nums, 0, func(acc, n int) int { return acc + n })
	if sum != 21 {
		t.Errorf("Reduce = %d, 想要 21", sum)
	}
}

func TestMax(t *testing.T) {
	if got := Max(3, 9); got != 9 {
		t.Errorf("Max(3,9) = %d", got)
	}
	if got := Max(9, 3); got != 9 {
		t.Errorf("Max(9,3) = %d", got)
	}
	if got := Max(1.5, 0.5); got != 1.5 {
		t.Errorf("Max(1.5,0.5) = %v", got)
	}
}

func TestStack(t *testing.T) {
	t.Run("LIFO 顺序", func(t *testing.T) {
		var s Stack[int]
		s.Push(1)
		s.Push(2)
		s.Push(3)

		if s.Len() != 3 {
			t.Fatalf("Len = %d", s.Len())
		}
		for _, want := range []int{3, 2, 1} {
			got, ok := s.Pop()
			if !ok || got != want {
				t.Fatalf("Pop = (%d, %t), 想要 (%d, true)", got, ok, want)
			}
		}
	})

	t.Run("空栈返回零值+false，不 panic", func(t *testing.T) {
		var s Stack[string]
		got, ok := s.Pop()
		if ok || got != "" {
			t.Errorf("空栈 Pop = (%q, %t), 想要 (\"\", false)", got, ok)
		}
	})

	t.Run("Peek 不弹出", func(t *testing.T) {
		var s Stack[int]
		s.Push(7)
		if v, _ := s.Peek(); v != 7 {
			t.Errorf("Peek = %d", v)
		}
		if s.Len() != 1 {
			t.Errorf("Peek 之后 Len 应该还是 1，实际 %d", s.Len())
		}
	})

	t.Run("泛型支持任意类型", func(t *testing.T) {
		type Point struct{ X, Y int }
		var s Stack[Point]
		s.Push(Point{1, 2})
		p, _ := s.Pop()
		if p != (Point{1, 2}) {
			t.Errorf("Pop = %+v", p)
		}
	})
}

// 标准库 slices / maps 的用法
func TestStdlibSlicesMaps(t *testing.T) {
	d := []int{3, 1, 4, 1, 5}

	if got := slices.Max(d); got != 5 {
		t.Errorf("slices.Max = %d", got)
	}
	if got := slices.Min(d); got != 1 {
		t.Errorf("slices.Min = %d", got)
	}
	if !slices.Contains(d, 4) {
		t.Error("应该包含 4")
	}
	if got := slices.Index(d, 5); got != 4 {
		t.Errorf("Index(5) = %d, 想要 4", got)
	}

	// Clone 是深拷贝一层（元素是值时就是完整独立）
	cp := slices.Clone(d)
	cp[0] = 100
	if d[0] == 100 {
		t.Error("Clone 应该产生独立切片")
	}

	// 排序 + 去重
	slices.Sort(d)
	if !slices.IsSorted(d) {
		t.Errorf("Sort 之后应该有序: %v", d)
	}
	if got := slices.Compact(slices.Clone(d)); !slices.Equal(got, []int{1, 3, 4, 5}) {
		t.Errorf("Compact = %v", got)
	}
}

func ExampleMap() {
	got := Map([]int{1, 2, 3}, func(n int) string {
		return fmt.Sprintf("<%d>", n)
	})
	fmt.Println(got)
	// Output: [<1> <2> <3>]
}
