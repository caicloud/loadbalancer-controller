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

package v1alpha1

import "fmt"

var (
	// LabelKeyProxy for all loadbalancer proxies
	// loadbalancer.alpha.caicloud.io/proxy
	LabelKeyProxy = fmt.Sprintf("%s.%s/%s", LoadBalancerName, AlphaGroupName, "proxy")

	// LabelKeyProvider for all loadbalancer providers
	// loadbalancer.alpha.caicloud.io/provider
	LabelKeyProvider = fmt.Sprintf("%s.%s/%s", LoadBalancerName, AlphaGroupName, "provider")

	// LabelKeyCreatedBy - loadbalancer.alpha.caicloud.io/created-by
	LabelKeyCreatedBy = fmt.Sprintf("%s.%s/created-by", LoadBalancerName, AlphaGroupName)
	// LabelValueFormatCreateby - namespace.name
	LabelValueFormatCreateby = "%s.%s"

	// UniqueLabelKeyFormat ...
	// loadbalancer.alpha.caicloud.io/namespace.name
	UniqueLabelKeyFormat = LoadBalancerName + "." + AlphaGroupName + "/" + "%s.%s"

	// TaintKey ...
	TaintKey = fmt.Sprintf("%s.%s/dedicated", LoadBalancerName, AlphaGroupName)
	// TaintValueFormat - namespace.name
	TaintValueFormat = "%s.%s"
)
