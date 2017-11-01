/*
Copyright 2017 caicloud authors. All rights reserved.
*/

package syncqueue

import (
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/golang/glog"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

const (
	maxRetries = 3
)

var (
	// defaultKeyFunc is the default key function
	defaultKeyFunc = cache.DeletionHandlingMetaNamespaceKeyFunc
)

// SyncHandler describes the function which will be
// called for each item in the queue
type SyncHandler func(obj interface{}) error

// KeyFunc describes a function that generates a key from a object
type KeyFunc func(obj interface{}) (interface{}, error)

// PassthroughKeyFunc is a keyFunc which returns the original obj
func PassthroughKeyFunc(obj interface{}) (interface{}, error) {
	return obj, nil
}

// SyncQueue is a helper for creating a kubernetes controller easily
// It requires a syncHandler and an optional key function.
// After running the syncQueue, you can call it's Enqueque function to enqueue items.
// SyncQueue will get key from the items by keyFunc, and add the key to the rate limit workqueue.
// The worker will be invoked to call the syncHandler.
type SyncQueue struct {
	// syncType is the object type in the queue
	syncType reflect.Type
	// queue is the work queue the worker polls
	queue workqueue.RateLimitingInterface
	// syncHandler is called for each item in the queue
	syncHandler SyncHandler
	// KeyFunc is called to get key from obj
	keyFunc KeyFunc

	waitGroup sync.WaitGroup

	maxRetries int
	stopCh     chan struct{}
}

// NewSyncQueue returns a new SyncQueue, enqueue key of obj using default keyFunc
func NewSyncQueue(syncObject runtime.Object, syncHandler SyncHandler) *SyncQueue {
	return NewCustomSyncQueue(syncObject, syncHandler, nil)
}

// NewPassthroughSyncQueue returns a new SyncQueue with PassthroughKeyFunc
func NewPassthroughSyncQueue(syncObject runtime.Object, syncHandler SyncHandler) *SyncQueue {
	return NewCustomSyncQueue(syncObject, syncHandler, PassthroughKeyFunc)
}

// NewCustomSyncQueue returns a new SyncQueue using custom keyFunc
func NewCustomSyncQueue(syncObject runtime.Object, syncHandler SyncHandler, keyFunc KeyFunc) *SyncQueue {
	sq := &SyncQueue{
		syncType:    reflect.TypeOf(syncObject),
		queue:       workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter()),
		syncHandler: syncHandler,
		keyFunc:     keyFunc,
		waitGroup:   sync.WaitGroup{},
		maxRetries:  maxRetries,
		stopCh:      make(chan struct{}),
	}

	if keyFunc == nil {
		sq.keyFunc = sq.defaultKeyFunc
	}

	return sq
}

// Run starts n workers to sync
func (sq *SyncQueue) Run(workers int) {
	for i := 0; i < workers; i++ {
		go wait.Until(sq.worker, time.Second, sq.stopCh)
	}
}

// ShutDown shuts down the work queue and waits for the worker to ACK
func (sq *SyncQueue) ShutDown() {
	close(sq.stopCh)
	// sq shutdown the queue, then worker can't get key from queue
	// processNextWorkItem return false, and then waitGroup -1
	sq.queue.ShutDown()
	sq.waitGroup.Wait()
}

// IsShuttingDown returns if the method Shutdown was invoked
func (sq *SyncQueue) IsShuttingDown() bool {
	return sq.queue.ShuttingDown()
}

// SetMaxRetries sets the max retry times of the queue
func (sq *SyncQueue) SetMaxRetries(max int) {
	if max > 0 {
		sq.maxRetries = max
	}
}

// Queue returns the rate limit work queue
func (sq *SyncQueue) Queue() workqueue.RateLimitingInterface {
	return sq.queue
}

// Enqueue wraps queue.Add
func (sq *SyncQueue) Enqueue(obj interface{}) {

	if sq.IsShuttingDown() {
		return
	}

	key, err := sq.keyFunc(obj)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("Couldn't get key for %v %#v: %v", sq.syncType, obj, err))
		return
	}
	sq.queue.Add(key)
}

// EnqueueRateLimited wraps queue.AddRateLimited. It adds an item to the workqueue
// after the rate limiter says its ok
func (sq *SyncQueue) EnqueueRateLimited(obj interface{}) {

	if sq.IsShuttingDown() {
		return
	}

	key, err := sq.keyFunc(obj)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("Couldn't get key for %v %#v: %v", sq.syncType, obj, err))
		return
	}
	sq.queue.AddRateLimited(key)
}

// EnqueueAfter wraps queue.AddAfter. It adds an item to the workqueue after the indicated duration has passed
func (sq *SyncQueue) EnqueueAfter(obj interface{}, after time.Duration) {

	if sq.IsShuttingDown() {
		return
	}

	key, err := sq.keyFunc(obj)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("Couldn't get key for %v %#v: %v", sq.syncType, obj, err))
		return
	}
	sq.queue.AddAfter(key, after)
}

func (sq *SyncQueue) defaultKeyFunc(obj interface{}) (interface{}, error) {
	key, err := defaultKeyFunc(obj)
	if err != nil {
		return "", err
	}
	return key, nil
}

// Worker is a common worker for controllers
// worker runs a worker thread that just dequeues items, processes them, and marks them done.
// It enforces that the syncHandler is never invoked concurrently with the same key.
func (sq *SyncQueue) worker() {
	sq.waitGroup.Add(1)
	defer sq.waitGroup.Done()
	// invoked oncely process any until exhausted
	for sq.processNextWorkItem() {
	}
}

// ProcessNextWorkItem processes next item in queue by syncHandler
func (sq *SyncQueue) processNextWorkItem() bool {
	obj, quit := sq.queue.Get()
	if quit {
		return false
	}
	defer sq.queue.Done(obj)

	err := sq.syncHandler(obj)
	sq.handleSyncError(err, obj)

	return true
}

// HandleSyncError handles error when sync obj error and retry n times
func (sq *SyncQueue) handleSyncError(err error, obj interface{}) {
	if err == nil {
		// no err
		sq.queue.Forget(obj)
		return
	}

	var key interface{}

	// get short key no matter what the keyfunc is
	key, kerr := defaultKeyFunc(obj)
	if kerr != nil {
		key = obj
	}

	if sq.queue.NumRequeues(obj) < sq.maxRetries {
		glog.Warningf("Error syncing object (type: %v, key: %v) retry: %v, err: %v",
			sq.syncType, key, sq.queue.NumRequeues(obj), err)

		sq.queue.AddRateLimited(obj)
		return
	}

	utilruntime.HandleError(err)
	glog.Warningf("Dropping object (type: %v, key: %v) from the queue, err: %v", sq.syncType, key, err)
	sq.queue.Forget(obj)
}
