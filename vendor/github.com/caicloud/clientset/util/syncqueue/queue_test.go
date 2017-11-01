/*
Copyright 2017 caicloud authors. All rights reserved.
*/

package syncqueue

import (
	"errors"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/pkg/api/v1"
)

func TestSyncQueue_Enqueue(t *testing.T) {

	syncPods := func(obj interface{}) error {
		pod := obj.(*v1.Pod)
		pod.Name = pod.Name + "_synced"
		return nil
	}

	queue := NewCustomSyncQueue(&v1.Pod{}, syncPods, PassthroughKeyFunc)
	queue.Run(1)
	defer func() {
		queue.ShutDown()
	}()

	tests := []struct {
		podName string
		want    string
	}{
		{"test", "test_synced"},
	}
	for _, tt := range tests {
		pod := &v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name: tt.podName,
			},
		}
		queue.Enqueue(pod)

		time.Sleep(1 * time.Millisecond)
		if pod.Name != tt.want {
			t.Errorf("SyncQueque.Enqueque() == %v, want %v", pod.Name, tt.want)
		}

	}
}

func TestSyncQueue_EnqueueError(t *testing.T) {
	syncError := func(obj interface{}) error {
		pod := obj.(*v1.Pod)
		if pod.Name == "test" {
			pod.Name = "test_1"
			return errors.New("error")
		}
		if pod.Name == "test_1" {
			pod.Name = "test_synced"
			return nil
		}
		return nil
	}
	queue := NewCustomSyncQueue(&v1.Pod{}, syncError, PassthroughKeyFunc)
	queue.SetMaxRetries(1)
	queue.Run(1)
	defer func() {
		queue.ShutDown()
	}()

	tests := []struct {
		podName string
		want    string
	}{
		{"test", "test_synced"},
	}
	for _, tt := range tests {
		pod := &v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name: tt.podName,
			},
		}
		queue.Enqueue(pod)

		time.Sleep(10 * time.Millisecond)
		if pod.Name != tt.want {
			t.Errorf("SyncQueque.Enqueque() == %v, want %v", pod.Name, tt.want)
		}

	}
}
