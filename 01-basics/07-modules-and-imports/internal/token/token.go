// Package token 放在 internal/ 下面，用来演示 Go 的可见性规则。
//
// **internal 规则**：路径里含有 internal 的那一层，它的父目录及其子树可以导入这个包，
// 其他任何地方都不行（编译器强制）。
//
// 这里 internal 的父目录是 01-basics/07-modules-and-imports，
// 所以：
//
//	✓ 01-basics/07-modules-and-imports/...
//	✗ 01-basics/01-hello/...
//	✗ 02-concurrency/...
//	✗ 任何外部模块
//
// 为什么要有这个机制？因为 Go 没有 private 关键字来限制"包之间的可见性"。
// 一个包一旦导出，就等于对全世界承诺不再改。
// internal 让你可以放心重构"内部实现"，而不用担心有人在别的地方依赖它。
package token

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
)

// ErrEmpty 空输入错误。
var ErrEmpty = errors.New("输入不能为空")

// Sign 计算输入的 SHA-256，返回十六进制字符串。
// 这是个纯函数，没有任何隐藏状态，测试起来非常容易。
func Sign(s string) (string, error) {
	if s == "" {
		return "", ErrEmpty
	}
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:]), nil
}

// MustSign 是 Sign 的"非空保证"版本，输入为空直接 panic。
// 命名惯例：MustXxx 表示"失败会 panic"。
func MustSign(s string) string {
	out, err := Sign(s)
	if err != nil {
		panic(fmt.Sprintf("token.MustSign: %v", err))
	}
	return out
}

// Random 生成 n 字节的随机十六进制串。
func Random(n int) (string, error) {
	if n <= 0 {
		return "", fmt.Errorf("token.Random: 长度必须为正数，实际 %d", n)
	}
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("token.Random: 读取随机源失败: %w", err)
	}
	return hex.EncodeToString(buf), nil
}
