# x-zoot

这是一个独立的 Go module，可以单独添加第三方依赖并运行。

运行：

    make run p=projects/x-zoot

测试：

    make test p=projects/x-zoot

该项目拥有自己的 `go.mod`，不会把依赖写入仓库根模块。
