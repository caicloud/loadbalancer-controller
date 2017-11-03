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

	lbapi "github.com/caicloud/clientset/pkg/apis/loadbalance/v1alpha2"
)

func TestProxyStatusEqual(t *testing.T) {

	tests := []struct {
		a    lbapi.ProxyStatus
		b    lbapi.ProxyStatus
		want bool
	}{
		{
			lbapi.ProxyStatus{},
			lbapi.ProxyStatus{},
			true,
		},
		{
			lbapi.ProxyStatus{
				Deployment: "test",
				ConfigMap:  "cm",
			},
			lbapi.ProxyStatus{
				Deployment: "test",
				ConfigMap:  "cm",
			},
			true,
		},
		{
			lbapi.ProxyStatus{
				Deployment: "test1",
			},
			lbapi.ProxyStatus{
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
		a    lbapi.PodStatuses
		b    lbapi.PodStatuses
		want bool
	}{
		{
			lbapi.PodStatuses{},
			lbapi.PodStatuses{},
			true,
		},
		{
			lbapi.PodStatuses{
				Replicas: 1,
				Statuses: []lbapi.PodStatus{
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
			lbapi.PodStatuses{
				Replicas: 1,
				Statuses: []lbapi.PodStatus{
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
