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

package toleration

import (
	lbapi "github.com/caicloud/clientset/pkg/apis/loadbalance/v1alpha2"
	"k8s.io/client-go/pkg/api/v1"
)

var (
	tolerationKeys = []string{lbapi.TaintKey}
)

// AddAdditionalTolerationKeys append additional toleration keys which loadbalancer should tolerate
func AddAdditionalTolerationKeys(keys []string) {
	tolerationKeys = append(tolerationKeys, keys...)
}

// GenerateTolerations generates the tolerations
func GenerateTolerations() []v1.Toleration {
	tolerations := make([]v1.Toleration, 0)

	for _, key := range tolerationKeys {
		tolerations = append(tolerations, v1.Toleration{
			Key:      key,
			Operator: v1.TolerationOpExists,
		})
	}

	return tolerations
}
