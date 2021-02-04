# Sync Queue

sync queue 简化了写 controller 时对 queue 的操作，减少代码量.

sync queue 可以用来代替 `workqueue.RatelimitInterface`，它接收一个 `SyncHandler` 和一个可选的 `KeyFunc`

在 Run 起来之后，使用方通过 `syncqueue.Enqueue`, 往队列里面添加 item，sync queue 使用 `KeyFunc` 从 item 中获取到 key，然后将 key 放入真正的 `workqueue` 中，`syncqueue` 的 worker 会从 `workqueue` 中获取 key，然后调用 `SyncHandler`，从而完成一次 controller business logic 的 loop

## Basic Usage

默认情况下， sync queue 使用`cache.DeletionHandlingMetaNamespaceKeyFunc` 做 `KeyFunc`，得到的 `key` 为 `namespace/name`，可以使用 `cache.SplitMetaNamespaceKey`，解析出 namespace 和 name

整个工作流程大致如下：

```go
func NewPodController() *PodController {
    c := &PodController{}

    c.queue = syncqueue.NewSyncQueue(&v1.Pod{}, c.sync)
    podinfomer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
        AddFunc: c.addPod,
    })
}

func (c *PodController) addPod(obj interface{}) {
    pod := obj.(*v1.Pod)
    c.queue.Enqueue(pod)
}

func (c *PodController) Run(workers int, stopCh <-chan struct{}) {
    // start informer
    // wait cache sync
    c.queue.Run(workers, stopCh)
    defer func() {
      c.queue.ShutDown()
    }
    <-stopCh
}

func (c *PodController) sync(obj interface{}) error {
    key := obj.(string)
    namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
    }
    
    pod, err := c.podLister.Pods(namespace).Get(name)
    if errors.IsNotFound(err) {
		// pod has been deleted
		return nil
	}
    if err != nil {
        return err
    }

    // business logic
    
}

```

## Custom Key Func

在上面例子中使用了默认的`KeyFunc`，下面说明如何使用自定义 `KeyFunc`

在这种模式下，除了创建 queue 的时候需要指定 keyFunc 外，在你的 syncHandler 中获取 obj 的方式也要做相应的修改

```go
func keyFunc(obj interface) (interface{}, err) {
    return obj, nil
}

func NewPodController() *PodController {
	...
    c.queue = syncqueue.NewCustomSyncQueue(&v1.Pod{}, c.sync, keyFunc)
    ...
}

func (c *PodController) sync(obj interface{}) error {
    pod := obj.(*v1.Pod)
    ...
}
```

