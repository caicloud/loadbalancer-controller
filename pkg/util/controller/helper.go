/*
Copyright 2017 Caicloud authors. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"fmt"
	"reflect"
	"time"

	log "github.com/zoumo/logdog"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

const (
	maxRetries = 3
)

var (
	// KeyFunc ...
	KeyFunc = cache.DeletionHandlingMetaNamespaceKeyFunc
)

// SyncHandler ...
type SyncHandler func(key interface{}) error

// Helper is a helper for creating a k8s controller easily
type Helper struct {
	SyncType reflect.Type
	Queue    workqueue.RateLimitingInterface

	SyncHandler         SyncHandler
	Enqueue             func(obj interface{})
	EnqueueRateLimited  func(obj interface{})
	EnqueueAfter        func(obj interface{}, after time.Duration)
	ProcessNextWorkItem func() bool
}

// NewHelper returns a new helper, enqueue key of obj
func NewHelper(syncObject runtime.Object, queue workqueue.RateLimitingInterface, syncHandler SyncHandler) *Helper {
	helper := &Helper{
		SyncType:    reflect.TypeOf(syncObject),
		Queue:       queue,
		SyncHandler: syncHandler,
	}

	helper.Enqueue = helper.enqueue
	helper.EnqueueRateLimited = helper.enqueueRateLimited
	helper.EnqueueAfter = helper.enqueueAfter
	helper.ProcessNextWorkItem = helper.processNextWorkItem

	return helper
}

// NewHelperForObj returns a new helper, enqueue whole obj
func NewHelperForObj(syncObject runtime.Object, queue workqueue.RateLimitingInterface, syncHandler SyncHandler) *Helper {
	helper := &Helper{
		SyncType:    reflect.TypeOf(syncObject),
		Queue:       queue,
		SyncHandler: syncHandler,
	}

	helper.Enqueue = helper.Queue.Add
	helper.EnqueueRateLimited = helper.Queue.AddRateLimited
	helper.EnqueueAfter = helper.Queue.AddAfter
	helper.ProcessNextWorkItem = helper.processNextWorkItem

	return helper
}

// Enqueue wraps queue.Add
func (helper *Helper) enqueue(obj interface{}) {
	key, err := KeyFunc(obj)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("Couldn't get key for %v %#v: %v", helper.SyncType, obj, err))
		return
	}
	helper.Queue.Add(key)
}

// EnqueueRateLimited wraps queue.AddRateLimited. It adds an item to the workqueue
// after the rate limiter says its ok
func (helper *Helper) enqueueRateLimited(obj interface{}) {
	key, err := KeyFunc(obj)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("Couldn't get key for %v %#v: %v", helper.SyncType, obj, err))
		return
	}
	helper.Queue.AddRateLimited(key)
}

// EnqueueAfter wraps queue.AddAfter. It adds an item to the workqueue after the indicated duration has passed
func (helper *Helper) enqueueAfter(obj interface{}, after time.Duration) {
	key, err := KeyFunc(obj)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("Couldn't get key for %v %#v: %v", helper.SyncType, obj, err))
		return
	}
	helper.Queue.AddAfter(key, after)
}

// Worker is a common worker for controllers
// worker runs a worker thread that just dequeues items, processes them, and marks them done.
// It enforces that the syncHandler is never invoked concurrently with the same key.
func (helper *Helper) Worker() {
	// invoked oncely process any until exhausted
	for helper.ProcessNextWorkItem() {
	}
}

// ProcessNextWorkItem process next item in queue by syncHandler
func (helper *Helper) processNextWorkItem() bool {
	obj, quit := helper.Queue.Get()
	if quit {
		return false
	}
	defer helper.Queue.Done(obj)

	err := helper.SyncHandler(obj)
	helper.HandleSyncError(err, obj)

	return true
}

// HandleSyncError handles error when sync obj error and retry n times
func (helper *Helper) HandleSyncError(err error, obj interface{}) {
	if err == nil {
		// no err
		helper.Queue.Forget(obj)
		return
	}

	var key interface{}
	key, kerr := KeyFunc(obj)
	if kerr != nil {
		key = obj
	}

	if helper.Queue.NumRequeues(obj) < maxRetries {
		log.Warn("Error syncing object, retry", log.Fields{"type": helper.SyncType, "obj": key, "err": err})
		helper.Queue.AddRateLimited(obj)
		return
	}

	utilruntime.HandleError(err)
	log.Warn("Dropping object out of queue", log.Fields{"type": helper.SyncType, "obj": key, "err": err})
	helper.Queue.Forget(obj)
}
