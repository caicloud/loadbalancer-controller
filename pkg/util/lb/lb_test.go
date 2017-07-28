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

package lb

import (
	"testing"

	netv1alpha1 "github.com/caicloud/loadbalancer-controller/pkg/apis/networking/v1alpha1"
)

func TestProxyStatusEqual(t *testing.T) {

	tests := []struct {
		a    netv1alpha1.ProxyStatus
		b    netv1alpha1.ProxyStatus
		want bool
	}{
		{
			netv1alpha1.ProxyStatus{},
			netv1alpha1.ProxyStatus{},
			true,
		},
		{
			netv1alpha1.ProxyStatus{
				Deployment: "test",
				ConfigMap:  "cm",
			},
			netv1alpha1.ProxyStatus{
				Deployment: "test",
				ConfigMap:  "cm",
			},
			true,
		},
		{
			netv1alpha1.ProxyStatus{
				Deployment: "test1",
			},
			netv1alpha1.ProxyStatus{
				Deployment: "test2",
			},
			false,
		},
	}
	for _, tt := range tests {
		if got := ProxyStatusEqual(tt.a, tt.b); got != tt.want {
			t.Errorf("ProxyStatusEqual() = %v, want %v, a: %v b: %v", got, tt.want, tt.a, tt.b)
		}
	}
}

func TestPodStatusesEqual(t *testing.T) {

	tests := []struct {
		a    netv1alpha1.PodStatuses
		b    netv1alpha1.PodStatuses
		want bool
	}{
		{
			netv1alpha1.PodStatuses{},
			netv1alpha1.PodStatuses{},
			true,
		},
		{
			netv1alpha1.PodStatuses{
				Replicas: 1,
				Statuses: []netv1alpha1.PodStatus{
					{
						Name:  "test1",
						Ready: true,
					},
					{
						Name:  "test2",
						Ready: true,
					},
				},
			},
			netv1alpha1.PodStatuses{
				Replicas: 1,
				Statuses: []netv1alpha1.PodStatus{
					{
						Name:  "test2",
						Ready: true,
					},
					{
						Name:  "test1",
						Ready: true,
					},
				},
			},
			true,
		},
	}
	for _, tt := range tests {
		if got := PodStatusesEqual(tt.a, tt.b); got != tt.want {
			t.Errorf("ProxyStatusEqual() = %v, want %v, a: %v b: %v", got, tt.want, tt.a, tt.b)
		}
	}
}
