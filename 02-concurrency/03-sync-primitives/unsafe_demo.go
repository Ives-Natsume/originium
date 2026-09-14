//go:build ignore

// 这个文件带 `//go:build ignore` 标签，所以**默认不参与构建**，
// `go build ./...` / `go test ./...` 都不会碰它。
//
// 它的作用是让你亲手看到"数据竞争"长什么样。
//
// 运行方式（注意要单独指定文件名，不能按包路径跑）：
//
//	go run 02-concurrency/03-sync-primitives/unsafe_demo.go
//	go run -race 02-concurrency/03-sync-primitives/unsafe_demo.go
//
// 第二个命令会打印出完整的竞态报告，包括：
//   - 哪个变量被竞争（地址）
//   - 两个 goroutine 各自的读写位置（文件:行号）
//   - 创建这些 goroutine 的调用栈
//
// 学会读这个报告，是排查并发 bug 的核心技能。
package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

func main() {
	const n = 1000

	// ---------- 错误示范：无保护 ----------
	var unsafeCount int
	var wg sync.WaitGroup
	for range n {
		wg.Add(1)
		go func() {
			defer wg.Done()
			unsafeCount++ // ← 数据竞争：读-改-写不是原子的
		}()
	}
	wg.Wait()

	fmt.Printf("无保护计数器: %d / %d", unsafeCount, n)
	if unsafeCount != n {
		fmt.Printf("  ← 丢了 %d 次更新！\n", n-unsafeCount)
	} else {
		fmt.Println("  ← 这次侥幸没丢，但这是不确定的（竞态就是这样时隐时现）")
	}

	// ---------- 正确做法 1：atomic ----------
	var atomicCount atomic.Int64
	for range n {
		wg.Add(1)
		go func() {
			defer wg.Done()
			atomicCount.Add(1)
		}()
	}
	wg.Wait()
	fmt.Printf("atomic 计数器: %d / %d  ← 永远正确\n", atomicCount.Load(), n)

	// ---------- 正确做法 2：Mutex ----------
	var (
		mu         sync.Mutex
		mutexCount int
	)
	for range n {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mu.Lock()
			mutexCount++
			mu.Unlock()
		}()
	}
	wg.Wait()
	fmt.Printf("Mutex 计数器:  %d / %d  ← 永远正确\n", mutexCount, n)

	fmt.Println("\n用 -race 再跑一次，你会看到无保护那段的完整竞态报告：")
	fmt.Println("  go run -race 02-concurrency/03-sync-primitives/unsafe_demo.go")
}
