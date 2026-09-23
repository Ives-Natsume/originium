# Originium · Go 学习工作台

这个仓库同时支持两种 Go 代码组织方式。

## 根模块练习

`01-basics/`、`02-concurrency/` 等目录属于根模块 `github.com/Ives-Natsume/originium`。
它们适合学习语言基础、包、导入和并发，继续使用原有命令：

```text
make run p=01-basics/01-hello
make test p=02-concurrency/01-goroutine-basics
make test-race p=02-concurrency/03-sync-primitives
```

## 独立完整子项目

需要使用 Gin、数据库或其他第三方依赖的完整应用，建议放在 `projects/` 下，每个项目拥有自己的 `go.mod`：

```text
projects/
  gin-service/
    go.mod
    main.go
    main_test.go
```

创建项目骨架：

```text
make project p=projects/gin-service
```

项目内添加依赖时，在项目目录执行 Go module 命令，依赖会写入该项目的 `go.mod` 和 `go.sum`，不会污染根模块：

```text
cd projects/gin-service
go get github.com/gin-gonic/gin
go mod tidy
```

仓库命令会根据目标路径自动找到最近的 `go.mod`：

```text
make run p=projects/gin-service
make test p=projects/gin-service
make check p=projects/gin-service
make fmt p=projects/gin-service
make tidy p=projects/gin-service
```

不指定 `p` 时，`make build`、`make test` 和 `make check` 会逐个检查仓库内的所有 Go module。

## 模块边界

脚手架以目录中的 `go.mod` 作为独立项目边界，不依赖目录名约定，也不要求提交 `go.work`。每个项目应该在 `GOWORK=off` 下可以独立构建和测试。

现有的 `x-web-service-gin/` 仍是学习中的实验目录。本次脚手架扩展不会修改其中的业务代码或自动补充 Gin 依赖；如果要正式作为独立 Gin 项目，可以使用 `make project` 创建新项目后再迁移练习内容。
