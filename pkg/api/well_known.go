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

package api

import (
	lbapi "github.com/caicloud/clientset/pkg/apis/loadbalance/v1alpha2"
)

const (
	// KeyStatic ...
	KeyStatic = lbapi.GroupName + "/static"

	// LabelHostname ...
	LabelHostname = "kubernetes.io/hostname"

	// LoadBalancerKind ...
	LoadBalancerKind = "LoadBalancer"
)

var (
	// ControllerKind contains the schema.GroupVersionKind for this controller type.
	ControllerKind = lbapi.SchemeGroupVersion.WithKind(LoadBalancerKind)
)
