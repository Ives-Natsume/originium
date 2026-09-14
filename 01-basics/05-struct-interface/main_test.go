package main

import (
	"fmt"
	"math"
	"sort"
	"testing"
)

const eps = 1e-9

func almostEqual(a, b float64) bool { return math.Abs(a-b) < eps }

func TestRect(t *testing.T) {
	r := Rect{W: 3, H: 4}
	if got := r.Area(); got != 12 {
		t.Errorf("Area() = %v, 想要 12", got)
	}
	if got := r.Perimeter(); got != 14 {
		t.Errorf("Perimeter() = %v, 想要 14", got)
	}
	if got := r.Name(); got != "矩形" {
		t.Errorf("Name() = %q", got)
	}
}

func TestCircle(t *testing.T) {
	c := Circle{R: 2}
	if got := c.Area(); !almostEqual(got, 4*math.Pi) {
		t.Errorf("Area() = %v, 想要 %v", got, 4*math.Pi)
	}
	if got := c.Perimeter(); !almostEqual(got, 4*math.Pi) {
		t.Errorf("Perimeter() = %v, 想要 %v", got, 4*math.Pi)
	}
}

// 这是"接口多态"的核心测试：同一个循环处理不同具体类型。
func TestShapeInterface(t *testing.T) {
	shapes := []Shape{Rect{W: 2, H: 3}, Circle{R: 1}}

	// 断言具体类型
	if _, ok := shapes[0].(Rect); !ok {
		t.Error("shapes[0] 应该是 Rect")
	}
	if _, ok := shapes[1].(*Rect); ok {
		t.Error("shapes[1] 不该是 *Rect")
	}

	sum := 0.0
	for _, s := range shapes {
		sum += s.Area()
	}
	if !almostEqual(sum, 6+math.Pi) {
		t.Errorf("面积和 = %v, 想要 %v", sum, 6+math.Pi)
	}
}

func TestDescribe(t *testing.T) {
	tests := []struct {
		name string
		in   any
		want string
	}{
		{"nil", nil, "空值"},
		{"整数", 42, "整数 42"},
		{"字符串", "abc", `字符串 "abc"（长度 3）`},
		{"中文长度按字符算", "你好", `字符串 "你好"（长度 2）`},
		{"布尔", true, "布尔 true"},
		{"实现 Shape 的结构体", Rect{W: 2, H: 5}, "矩形，面积 10.00"},
		{"未实现 Stringer 的切片", []int{1}, "未知类型 []int"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Describe(tt.in); got != tt.want {
				t.Errorf("Describe(%v) = %q, 想要 %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestShapeInfo(t *testing.T) {
	t.Run("满足接口时 ok=true", func(t *testing.T) {
		info, ok := ShapeInfo(Circle{R: 1})
		if !ok {
			t.Fatal("Circle 应该满足 Shape")
		}
		_ = info
	})

	t.Run("不满足接口时 ok=false 且不 panic", func(t *testing.T) {
		if _, ok := ShapeInfo("字符串"); ok {
			t.Error("字符串不该满足 Shape")
		}
		if _, ok := ShapeInfo(nil); ok {
			t.Error("nil 不该满足 Shape")
		}
	})
}

// 指针接收者的方法**不会**被值类型满足。
// 这是新手最容易踩的坑，所以专门测一下。
func TestMethodSet(t *testing.T) {
	c := Counter{}
	c.Inc()
	c.Inc()
	c.Add(3)

	if got := c.Value(); got != 5 {
		t.Errorf("Value() = %d, 想要 5", got)
	}

	// 因为 Inc/Add 是指针接收者，所以只有 *Counter 满足 fmt.Stringer
	var s fmt.Stringer = &c
	if got := s.String(); got != "Counter(5)" {
		t.Errorf("String() = %q", got)
	}

	// 值拷贝不会互相影响
	cp := c
	cp.Inc()
	if c.Value() == cp.Value() {
		t.Error("拷贝后修改不该影响原对象")
	}
}

// 嵌入：外层同名方法覆盖内层，未覆盖的则被"提升"。
func TestEmbedding(t *testing.T) {
	d := Dog{Animal: Animal{Name: "旺财", Legs: 4}, Breed: "柴犬"}

	// 提升的字段
	if d.Name != "旺财" {
		t.Errorf("d.Name = %q，嵌入字段应被提升", d.Name)
	}
	if d.Animal.Name != d.Name {
		t.Error("d.Name 应该就是 d.Animal.Name")
	}

	// 覆盖的方法
	if got := d.Speak(); got != "旺财 汪汪叫" {
		t.Errorf("Dog.Speak() = %q", got)
	}
	// 显式穿透到被覆盖的版本
	if got := d.Animal.Speak(); got != "旺财 发出了声音" {
		t.Errorf("Animal.Speak() = %q", got)
	}
	// 提升的方法
	if got := d.Describe(); got != "旺财 有 4 条腿" {
		t.Errorf("Describe() = %q", got)
	}

	// Cat 没有覆盖 Speak
	c := Cat{Animal: Animal{Name: "咪咪", Legs: 4}}
	if got := c.Speak(); got != "咪咪 发出了声音" {
		t.Errorf("Cat 用的是 Animal 的 Speak，实际 %q", got)
	}
}

func TestSortByArea(t *testing.T) {
	shapes := []Shape{
		Rect{W: 10, H: 10}, // 100
		Circle{R: 1},       // 3.14
		Rect{W: 2, H: 2},   // 4
	}
	sort.Sort(ByArea(shapes))

	for i := 1; i < len(shapes); i++ {
		if shapes[i-1].Area() > shapes[i].Area() {
			t.Fatalf("排序失败: 第 %d 项比第 %d 项大", i-1, i)
		}
	}
	if shapes[0].Area() > 4.0 {
		t.Errorf("最小面积应该是圆(约 3.14)，实际 %v", shapes[0].Area())
	}
}

func TestSortSlicePeople(t *testing.T) {
	people := []Person{{"Neo", 30}, {"李雷", 25}, {"Alex", 30}, {"Bob", 41}}
	sort.Slice(people, func(i, j int) bool {
		if people[i].Age != people[j].Age {
			return people[i].Age < people[j].Age
		}
		return people[i].Name < people[j].Name
	})

	want := []Person{{"李雷", 25}, {"Alex", 30}, {"Neo", 30}, {"Bob", 41}}
	for i := range want {
		if people[i] != want[i] {
			t.Errorf("第 %d 个 = %+v, 想要 %+v", i, people[i], want[i])
		}
	}
}

func TestFormatAll(t *testing.T) {
	got := FormatAll([]Shape{Rect{W: 2, H: 3}, Circle{R: 1}})
	const want = "汇总: 矩形(6.0), 圆(3.1)"
	if got != want {
		t.Errorf("FormatAll() = %q, 想要 %q", got, want)
	}
}

// Example 验证的是"输出"，顺带当文档用。
func ExampleDescribe() {
	fmt.Println(Describe(Rect{W: 2, H: 5}))
	fmt.Println(Describe(42))
	// Output:
	// 矩形，面积 10.00
	// 整数 42
}
