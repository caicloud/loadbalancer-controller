package api

import "fmt"

var (
	// LabelKeyProxy for all loadbalancer proxies
	// loadbalancer.alpha.caicloud.io/proxy
	LabelKeyProxy = fmt.Sprintf("%s.%s/%s", LoadBalancerName, GroupName, "proxy")

	// LabelKeyProvider for all loadbalancer providers
	// loadbalancer.alpha.caicloud.io/provider
	LabelKeyProvider = fmt.Sprintf("%s.%s/%s", LoadBalancerName, GroupName, "provider")

	// LabelKeyCreateby - loadbalancer.alpha.caicloud.io/createby
	LabelKeyCreateby = fmt.Sprintf("%s.%s/createby", LoadBalancerName, GroupName)
	// LabelValueFormatCreateby - default_lbName
	LabelValueFormatCreateby = "%s_%s"

	// UniqueLabelKeyFormat ...
	// loadbalancer.alpha.caicloud.io/namespace-lbname
	UniqueLabelKeyFormat = LoadBalancerName + "." + GroupName + "/" + "%s_%s"

	// TaintKey ...
	TaintKey = fmt.Sprintf("%s.%s/dedicated", LoadBalancerName, GroupName)
	// TaintValueFormat ...
	TaintValueFormat = "%s_%s"
)
