package main

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestPingPong(t *testing.T) {
	got := PingPong(3)
	want := []string{"ping 0 -> pong 0", "ping 1 -> pong 1", "ping 2 -> pong 4"}
	if !slices.Equal(got, want) {
		t.Errorf("got %v, 想要 %v", got, want)
	}

	// 因为是无缓冲 channel，顺序是确定的 —— 这正是"同步"的价值
	if got := PingPong(0); len(got) != 0 {
		t.Errorf("rounds=0 应该返回空，实际 %v", got)
	}
}

func TestBufferedDemo(t *testing.T) {
	for _, cap := range []int{1, 3, 10} {
		got := BufferedDemo(cap)
		want := make([]int, cap)
		for i := range want {
			want[i] = i
		}
		if !slices.Equal(got, want) {
			t.Errorf("BufferedDemo(%d) = %v, 想要 %v", cap, got, want)
		}
	}
}

func TestCloseSemantics(t *testing.T) {
	values, closedVal, ok := CloseSemantics()

	if !slices.Equal(values, []int{1, 2}) {
		t.Errorf("range 应该读出 [1 2]，实际 %v", values)
	}
	if closedVal != 0 {
		t.Errorf("关闭后读到零值 0，实际 %d", closedVal)
	}
	if ok {
		t.Error("关闭后 ok 应该是 false")
	}
}

// 关闭之后不能发送 —— 这是 panic，所以要 recover 着测
func TestSendOnClosedChannelPanics(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("向已关闭的 channel 发送应该 panic")
		}
		if !strings.Contains(fmt.Sprint(r), "closed") {
			t.Errorf("panic 信息应该提到 closed，实际: %v", r)
		}
	}()

	ch := make(chan int, 1)
	close(ch)
	ch <- 1 // panic
}

func TestDoubleClosePanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("重复 close 应该 panic")
		}
	}()
	ch := make(chan int)
	close(ch)
	close(ch)
}

func TestFastest(t *testing.T) {
	url, d := Fastest([]string{"a", "b", "c"})
	// 第一个 goroutine 延迟 20ms，是最快的
	if url != "a" {
		t.Errorf("最快的应该是 a（延迟最小），实际 %q", url)
	}
	if d != 20*time.Millisecond {
		t.Errorf("延迟应该是 20ms，实际 %v", d)
	}

	// 单个也能工作
	if u, _ := Fastest([]string{"only"}); u != "only" {
		t.Errorf("got %q, 想要 only", u)
	}
}

func TestNonBlockingRecv(t *testing.T) {
	ch := make(chan int, 1)

	// 空的：立刻返回 false，不阻塞
	if v, ok := NonBlockingRecv(ch); ok || v != 0 {
		t.Errorf("空 channel 应该返回 (0, false)，实际 (%d, %t)", v, ok)
	}

	ch <- 7
	if v, ok := NonBlockingRecv(ch); !ok || v != 7 {
		t.Errorf("有数据时应该返回 (7, true)，实际 (%d, %t)", v, ok)
	}

	// 这个函数即使在无缓冲、无人接收的 channel 上调用也不会卡住
	unbuf := make(chan int)
	done := make(chan struct{})
	go func() { NonBlockingRecv(unbuf); close(done) }()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("NonBlockingRecv 不该阻塞")
	}
}

func TestNonBlockingSend(t *testing.T) {
	ch := make(chan int, 2)

	if !NonBlockingSend(ch, 1) || !NonBlockingSend(ch, 2) {
		t.Error("缓冲区有空间时应该发送成功")
	}
	if NonBlockingSend(ch, 3) {
		t.Error("缓冲区满了应该返回 false（丢弃）")
	}
}

func TestTimeout(t *testing.T) {
	t.Run("在时限内完成", func(t *testing.T) {
		got, err := Timeout(func() string {
			time.Sleep(5 * time.Millisecond)
			return "done"
		}, 200*time.Millisecond)
		if err != nil {
			t.Fatalf("不该超时: %v", err)
		}
		if got != "done" {
			t.Errorf("got %q", got)
		}
	})

	t.Run("超时返回错误", func(t *testing.T) {
		_, err := Timeout(func() string {
			time.Sleep(500 * time.Millisecond)
			return "too late"
		}, 10*time.Millisecond)
		if err == nil {
			t.Fatal("应该超时")
		}
		if !strings.Contains(err.Error(), "超时") {
			t.Errorf("错误信息应该说明超时: %v", err)
		}
	})
}

// pipeline 的端到端测试
func TestPipeline(t *testing.T) {
	gen := Generate(1, 2, 3, 4)
	sqd := make(chan int)
	go square(gen, sqd)

	filtered := make(chan int)
	go func() {
		defer close(filtered)
		for v := range sqd {
			if v > 5 {
				filtered <- v
			}
		}
	}()

	got := ToSlice(filtered)
	slices.Sort(got)
	if !slices.Equal(got, []int{9, 16}) {
		t.Errorf("got %v, 想要 [9 16]", got)
	}
}

func TestGenerate(t *testing.T) {
	t.Run("正常", func(t *testing.T) {
		got := ToSlice(Generate(1, 2, 3))
		if !slices.Equal(got, []int{1, 2, 3}) {
			t.Errorf("got %v", got)
		}
	})

	t.Run("空输入返回空切片而不是 nil channel", func(t *testing.T) {
		ch := Generate()
		got := ToSlice(ch) // 应该正常结束，不阻塞
		if len(got) != 0 {
			t.Errorf("got %v", got)
		}
	})
}

func TestFanIn(t *testing.T) {
	// fan-in 之后顺序不确定，所以要排序后比较
	got := ToSlice(FanIn(Generate(1, 3, 5), Generate(2, 4), Generate(6)))
	slices.Sort(got)

	want := []int{1, 2, 3, 4, 5, 6}
	if !slices.Equal(got, want) {
		t.Errorf("got %v, 想要 %v", got, want)
	}
}

func TestFanInEmptyOthers(t *testing.T) {
	// 有些输入是空的，fan-in 不能因此卡住（这是最容易写错的地方）
	got := ToSlice(FanIn(Generate(), Generate(1), Generate()))
	if !slices.Equal(got, []int{1}) {
		t.Errorf("got %v, 想要 [1]", got)
	}

	// 全是空的也要正常结束
	got = ToSlice(FanIn(Generate(), Generate()))
	if len(got) != 0 {
		t.Errorf("got %v", got)
	}
}

// ---------------------------------------------------------------------------
// 并发安全检查：worker 数量、并发读写
// ---------------------------------------------------------------------------

func TestFanInManyProducers(t *testing.T) {
	chans := make([]<-chan int, 50)
	for i := range chans {
		chans[i] = Generate(i)
	}
	got := ToSlice(FanIn(chans...))
	if len(got) != 50 {
		t.Errorf("期望 50 个元素，实际 %d", len(got))
	}
	slices.Sort(got)
	for i, v := range got {
		if v != i {
			t.Fatalf("第 %d 个 = %d，数据丢了", i, v)
		}
	}
}

func TestChannelAsSignal(t *testing.T) {
	// chan struct{} 是"纯信号"的惯用法：不传数据，只表示"发生了"
	done := make(chan struct{})
	var wg sync.WaitGroup

	for range 5 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-done // 所有 worker 在这里等同一个信号 -> 广播效果
		}()
	}

	close(done) // close 会唤醒所有等待者（发送只能唤醒一个）
	wg.Wait()   // 全部退出
}

func TestClosedChannelReadIsInstant(t *testing.T) {
	ch := make(chan int)
	close(ch)

	start := time.Now()
	for range 1000 {
		<-ch // 已关闭的 channel 读取永不阻塞
	}
	if elapsed := time.Since(start); elapsed > 100*time.Millisecond {
		t.Errorf("关闭的 channel 读取应该立刻返回，耗时 %v", elapsed)
	}
}

var _ = errors.New // 保留 errors 导入的占位
