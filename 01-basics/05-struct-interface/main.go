// Package main —— 第 5 课：结构体与接口。
//
// 运行: make run p=01-basics/05-struct-interface
package main

import (
	"fmt"
	"sort"
	"strings"
)

// ============================================================================
// 一、接口：由"使用方"定义，实现方不需要声明
// ============================================================================

// Shape 是一个接口。任何类型只要**拥有** Area() 和 Perimeter() 两个方法，
// 就自动满足它 —— 不需要写 "implements Shape"。
//
// 这是 Go 和其他语言最大的思路差异：
//
//	Java/C#：接口是"契约"，实现方主动声明
//	Go    ：接口是"能力描述"，使用方按需定义。实现方甚至可以不知道自己被用到了
type Shape interface {
	Area() float64
	Perimeter() float64
	Name() string
}

// ============================================================================
// 二、结构体与值/指针接收者
// ============================================================================

// Rect 矩形。面积用值接收者就够了 —— 它不修改 r。
type Rect struct {
	W, H float64
}

func (r Rect) Area() float64      { return r.W * r.H }
func (r Rect) Perimeter() float64 { return 2 * (r.W + r.H) }
func (r Rect) Name() string       { return "矩形" }

// Circle 圆形。
type Circle struct {
	R float64
}

func (c Circle) Area() float64      { return 3.141592653589793 * c.R * c.R }
func (c Circle) Perimeter() float64 { return 2 * 3.141592653589793 * c.R }
func (c Circle) Name() string       { return "圆" }

// Counter 演示**指针接收者**：要修改自身状态时必须用指针。
//
// 规则（记牢，这是新手最常踩的坑）：
//
//	值接收者  -> 方法集包含 T 和 *T
//	指针接收者 -> 方法集只包含 *T
//
// 所以：如果某个方法需要指针接收者，那这个类型的所有方法最好都用指针接收者，
// 保持一致性，否则 *T 和 T 能用的方法不一样，容易出迷惑 bug。
type Counter struct {
	n int
}

func (c *Counter) Inc()          { c.n++ }
func (c *Counter) Add(d int)     { c.n += d }
func (c Counter) Value() int     { return c.n }
func (c Counter) String() string { return fmt.Sprintf("Counter(%d)", c.n) }

// 编译期断言：如果 *Counter 不满足 fmt.Stringer，这行就编译不过。
// 这比等到运行时才发现要好得多 —— 是很实用的一个技巧。
var _ fmt.Stringer = (*Counter)(nil)
var _ Shape = Rect{}
var _ Shape = Circle{}

// ============================================================================
// 三、嵌入（embedding）：Go 的组合与"继承"
// ============================================================================

// Animal 被嵌入到别的类型里复用字段和方法。
type Animal struct {
	Name string
	Legs int
}

func (a Animal) Describe() string {
	return fmt.Sprintf("%s 有 %d 条腿", a.Name, a.Legs)
}

func (a Animal) Speak() string { return a.Name + " 发出了声音" }

// Dog 嵌入 Animal。
//
// 嵌入不是继承，是"自动获得一个匿名字段 + 它的方法被提升到外层"。
// 所以 d.Name 能直接用（其实是 d.Animal.Name 的语法糖）。
type Dog struct {
	Animal // 嵌入：字段名就是类型名 Animal
	Breed  string
}

// 外层同名方法会**覆盖**被提升的方法（类似方法重写，但没有多态分派）。
func (d Dog) Speak() string { return d.Name + " 汪汪叫" }

// Cat 不覆盖 Speak，所以直接用 Animal 提升上来的版本。
type Cat struct {
	Animal
	Indoor bool
}

// ============================================================================
// 四、类型断言与 type switch
// ============================================================================

// Describe 用 type switch 按实际类型分支处理。
// 参数是 any（= interface{}，Go 1.18+ 的别名）。
func Describe(v any) string {
	switch x := x2(v).(type) {
	case nil:
		return "空值"
	case int:
		return fmt.Sprintf("整数 %d", x)
	case string:
		return fmt.Sprintf("字符串 %q（长度 %d）", x, len([]rune(x)))
	case bool:
		return fmt.Sprintf("布尔 %t", x)
	case Shape: // 注意：接口 case 必须放在具体类型后面，语义是"实现了这个接口吗"
		return fmt.Sprintf("%s，面积 %.2f", x.Name(), x.Area())
	case fmt.Stringer:
		return "实现了 Stringer: " + x.String()
	default:
		return fmt.Sprintf("未知类型 %T", x)
	}
}

// x2 只是为了绕过编译器的类型推断，让 any 保持 any。
func x2(v any) any { return v }

// ShapeInfo 演示"接口 + ok 形式的断言"。
func ShapeInfo(v any) (string, bool) {
	s, ok := v.(Shape) // 断言失败时 ok=false，不会 panic
	if !ok {
		return "", false
	}
	return fmt.Sprintf("%s area=%.2f perimeter=%.2f", s.Name(), s.Area(), s.Perimeter()), true
}

// ============================================================================
// 五、排序
// ============================================================================

// ByArea 实现 sort.Interface，让 []Shape 可以排序。
type ByArea []Shape

func (s ByArea) Len() int           { return len(s) }
func (s ByArea) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s ByArea) Less(i, j int) bool { return s[i].Area() < s[j].Area() }

// Person 用来演示 sort.Slice（更常用，不用定义新类型）。
type Person struct {
	Name string
	Age  int
}

func main() {
	fmt.Println("===== 1. 接口与多态 =====")
	// 关键点：切片元素类型是接口，装的是不同具体类型
	shapes := []Shape{
		Rect{W: 3, H: 4},
		Circle{R: 1},
		Rect{W: 1, H: 1},
		Circle{R: 2},
	}
	total := 0.0
	for _, s := range shapes {
		fmt.Printf("%-4s 面积=%6.2f 周长=%6.2f\n", s.Name(), s.Area(), s.Perimeter())
		total += s.Area()
	}
	fmt.Printf("面积合计 = %.2f\n", total)

	fmt.Println("\n===== 2. 按面积排序（同一份数据，换个规则） =====")
	sort.Sort(ByArea(shapes)) // 从小到大
	for _, s := range shapes {
		fmt.Printf("%-4s %6.2f\n", s.Name(), s.Area())
	}

	fmt.Println("\n===== 3. sort.Slice：临时规则，不用定义新类型 =====")
	people := []Person{
		{"Neo", 30}, {"李雷", 25}, {"Alex", 30}, {"Bob", 41},
	}
	sort.Slice(people, func(i, j int) bool {
		if people[i].Age != people[j].Age {
			return people[i].Age < people[j].Age // 先按年龄
		}
		return people[i].Name < people[j].Name // 年龄相同按名字，保证结果稳定
	})
	for _, p := range people {
		fmt.Printf("%-6s %d\n", p.Name, p.Age)
	}
	// 还有 sort.SliceStable：排序后保持相等元素的原始相对顺序

	fmt.Println("\n===== 4. 值接收者 vs 指针接收者 =====")
	var counter Counter
	counter.Inc()
	counter.Add(10)
	fmt.Printf("counter.Value() = %d\n", counter.Value())

	// 直接打印结构体会调 String()
	fmt.Printf("直接打印: %v  （因为有 String() 方法）\n", &counter)

	// 这个坑要注意：值接收者方法拿到的是**拷贝**
	r := Rect{W: 2, H: 3}
	modifyRect(r) // 传值 -> 改不动外面的 r
	fmt.Printf("传值给函数后: W=%.0f（没变，因为是拷贝）\n", r.W)

	fmt.Println("\n===== 5. 嵌入 =====")
	d := Dog{Animal: Animal{Name: "旺财", Legs: 4}, Breed: "柴犬"}
	// 字段被"提升"：可以直接 d.Name 访问
	fmt.Printf("d.Name = %q（实际是 d.Animal.Name）\n", d.Name)
	fmt.Printf("d.Describe() = %q（提升自 Animal）\n", d.Describe())
	fmt.Printf("d.Speak()    = %q（Dog 自己覆盖了）\n", d.Speak())
	// 想调被覆盖的版本，得显式穿过嵌入字段
	fmt.Printf("d.Animal.Speak() = %q\n", d.Animal.Speak())
	fmt.Printf("显式访问嵌入字段: d.Animal.Name = %q\n", d.Animal.Name)

	c := Cat{Animal: Animal{Name: "咪咪", Legs: 4}, Indoor: true}
	fmt.Printf("c.Speak()    = %q（Cat 没覆盖，用的是 Animal 的）\n", c.Speak())

	fmt.Println("\n===== 6. 类型断言与 type switch =====")
	for _, v := range []any{
		nil, 42, "你好世界", true, Rect{W: 2, H: 5}, &counter,
		[]int{1, 2}, 3.14,
	} {
		fmt.Printf("  %s\n", Describe(v))
	}

	// ok 形式的断言：失败不 panic
	if info, ok := ShapeInfo(Rect{W: 2, H: 3}); ok {
		fmt.Printf("\nShapeInfo(Rect) = %s\n", info)
	}
	if _, ok := ShapeInfo("我不是 Shape"); !ok {
		fmt.Println("ShapeInfo(字符串) -> 断言失败，ok=false（没有 panic）")
	}

	fmt.Println("\n===== 7. 接口的另一种价值：依赖倒置 =====")
	fmt.Println(FormatAll(shapes))
}

// FormatAll 只依赖 Shape 接口，不关心具体是什么形状。
// 以后加三角形、五边形，这个函数一行都不用改。
func FormatAll(items []Shape) string {
	var b strings.Builder
	b.WriteString("汇总: ")
	for i, s := range items {
		if i > 0 {
			b.WriteString(", ")
		}
		fmt.Fprintf(&b, "%s(%.1f)", s.Name(), s.Area())
	}
	return b.String()
}

// modifyRect 演示值拷贝。
func modifyRect(r Rect) { r.W = 999 }
