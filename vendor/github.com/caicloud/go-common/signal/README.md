# signal

`SetupStopSignalHandler` 函数会返回一个 stop channel，当收到类似 ctrl-c 之类需要退出的 signal 时会自动 close 这个 stop channel；在 controller 和其它需要 stop channel 的 CLI 应用中非常有用，直接使用这个函数生成的 stop channel，不再需要其它的特别处理。下面有一个简单的示例：

```go
stopCh := signal.SetupStopSignalHandler()

fmt.Println("waiting stop signal")
<-stopCh
fmt.Println("stop signal received, exiting")
```

运行效果：

```
waiting stop signal
^Cstop signal received, exiting
```

没收到 stop signal 时，程序会一直卡在 `<-stopCh` 这个地方（在等待 stop signal），如果在程序退出前有额外的 clean up 工作需要做，在 stopCh 关闭后处理就行
