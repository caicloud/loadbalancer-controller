/*
Copyright 2017 caicloud authors. All rights reserved.
*/

package v1alpha2

import "fmt"

var (
	// LabelKeyProxy for all loadbalancer proxies
	// loadbalance.caicloud.io/proxy
	LabelKeyProxy = fmt.Sprintf("%s/%s", GroupName, "proxy")

	// LabelKeyProvider for all loadbalancer providers
	// loadbalance.caicloud.io/provider
	LabelKeyProvider = fmt.Sprintf("%s/%s", GroupName, "provider")

	// LabelKeyCreatedBy - loadbalance.caicloud.io/created-by
	LabelKeyCreatedBy = fmt.Sprintf("%s/created-by", GroupName)
	// LabelValueFormatCreateby - namespace.name
	LabelValueFormatCreateby = "%s.%s"

	// UniqueLabelKeyFormat - loadbalance.caicloud.io/namespace.name
	UniqueLabelKeyFormat = GroupName + "/" + "%s.%s"

	// TaintKey - loadbalance.caicloud.io/dedicated
	TaintKey = fmt.Sprintf("%s/dedicated", GroupName)
	// TaintValueFormat - namespace.name
	TaintValueFormat = "%s.%s"
)
