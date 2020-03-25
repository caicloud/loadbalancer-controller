<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Table of Contents**  *generated with [DocToc](https://github.com/thlorenz/doctoc)*

- [client](#client)
  - [1 快速上手](#1-%E5%BF%AB%E9%80%9F%E4%B8%8A%E6%89%8B)
  - [2 定制自己的 config](#2-%E5%AE%9A%E5%88%B6%E8%87%AA%E5%B7%B1%E7%9A%84-config)
    - [2.1 QPS](#21-qps)
    - [2.2 其它字段](#22-%E5%85%B6%E5%AE%83%E5%AD%97%E6%AE%B5)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

# client

在 client-go 中，用于构造 clientset 的 config 中存在 QPS 限制 ([见这里](https://github.com/kubernetes/client-go/blob/master/rest/config.go#L114))，各个组件的开发中也未考虑到这一点，导致访问 api-server 比较频繁的时候被限流。这个包对 clientset 的创建进行了封装，其内部为 config 设定了一些默认值，并对外提供了设置 QPS 的公用函数。相关讨论：

- [client-go 自带限流问题](https://github.com/caicloud/platform-arch/issues/47)
- [是否需要封装创建 kubernetes.Clientset 的函数](https://github.com/caicloud/go-common/issues/110)

## 1 快速上手

```golang
import (
    "github.com/caicloud/go-common/kubernetes/client"
)

func main() {
    // 集群内
    // client, err := client.NewFromFlags("", "")
    // 集群外
    client, err := client.NewFromFlags("", "the path of kubeconfig")
    if err != nil {
        panic(err)
    }
    ...
}
```

## 2 定制自己的 config

### 2.1 QPS

如果你确定需要修改 QPS 而不使用默认值（先向 TL 说明原因），可以使用公开的 `SetQPS(float32, int)` 函数：

```golang
import (
    "github.com/caicloud/go-common/kubernetes/client"
)

func main() {
    client, err := client.NewFromFlags("", "the path of kubeconfig", client.SetQPS(10, 20))
    if err != nil {
        panic(err)
    }
    ...
}
```

### 2.2 其它字段

可以自己写一个 `func(*rest.Config)` 的函数，传给 `NewFromFlags`函数：

```golang
import (
    "github.com/caicloud/go-common/kubernetes/client"

    "k8s.io/client-go/rest"
)

func main() {
    configModifier := func(c *rest.Config) {
        // 原地修改 c
        ...
    }
    client, err := client.NewFromFlags("", "the path of kubeconfig", configModifier)
    if err != nil {
        panic(err)
    }
    ...
}
```
