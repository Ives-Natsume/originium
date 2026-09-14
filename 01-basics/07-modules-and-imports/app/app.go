// Package app 演示"包依赖包"的层次结构。
//
// 依赖方向：main -> app -> {greeter, internal/token}
//
// 良好的包结构是**无环的**：上层依赖下层，下层绝不反过来引用上层。
// 一旦出现循环导入（import cycle），Go 会直接编译报错 ——
// 这其实是好事，它逼着你把公共部分抽到更底层的包里。
//
// 判断分层是否合理的一个简单标准：能不能用一句话说清每层的职责？
//
//	greeter  —— 生成问候语
//	token    —— 生成/校验令牌
//	app      —— 把上面两个组合成一个"业务流程"
//	main     —— 只负责解析参数、调用 app、处理退出码
package app

import (
	"fmt"
	"strings"

	"github.com/Ives-Natsume/originium/01-basics/07-modules-and-imports/greeter"
	"github.com/Ives-Natsume/originium/01-basics/07-modules-and-imports/internal/token"
)

// Welcome 是 app 包对外的主入口：给个名字，返回带令牌的欢迎语。
func Welcome(name string) (string, error) {
	if strings.TrimSpace(name) == "" {
		return "", fmt.Errorf("app.Welcome: 名字不能为空")
	}

	tok, err := token.Random(4)
	if err != nil {
		// 包边界是"错误翻译"的好位置：把底层错误包上本层的上下文
		return "", fmt.Errorf("app.Welcome: 生成令牌失败: %w", err)
	}

	return fmt.Sprintf("%s [token=%s]", greeter.Greet(name), tok), nil
}

// Describe 汇总各层的信息，方便在 main 里打印。
func Describe() []string {
	return []string{
		greeter.Stats(),
		fmt.Sprintf("greeter 版本 = %s", greeter.Version()),
		fmt.Sprintf("token(\"abc\") 摘要前 16 位 = %s", token.MustSign("abc")[:16]),
	}
}
