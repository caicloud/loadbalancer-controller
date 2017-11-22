package v1alpha2

import (
	"fmt"
	"net"
)

// ValidateLoadBalancer validate loadbalancer
func ValidateLoadBalancer(lb *LoadBalancer) error {

	// validate ipvsdr
	err := ValidateProviders(lb.Spec.Providers)
	if err != nil {
		return err
	}
	// validate proxy
	return ValidateProxy(lb.Spec.Proxy)
}

// ValidateProviders validate providers spec in loadbalancer
func ValidateProviders(spec ProvidersSpec) error {
	if spec.Ipvsdr != nil {
		ipvsdr := spec.Ipvsdr
		if net.ParseIP(ipvsdr.Vip) == nil {
			return fmt.Errorf("ipvsdr: vip is invalid")
		}
		switch ipvsdr.Scheduler {
		case IpvsSchedulerRR:
		case IpvsSchedulerWRR:
		case IpvsSchedulerLC:
		case IpvsSchedulerWLC:
		case IpvsSchedulerLBLC:
		case IpvsSchedulerDH:
		case IpvsSchedulerSH:
		default:
			return fmt.Errorf("ipvsdr: scheduler %v is invalid", ipvsdr.Scheduler)
		}
	}
	return nil
}

// ValidateProxy validate proxy spec in loadbalancer
func ValidateProxy(spec ProxySpec) error {
	switch spec.Type {
	case ProxyTypeNginx:
	case ProxyTypeHaproxy:
	case ProxyTypeTraefik:
	default:
		return fmt.Errorf("unknown proxy type %v", spec.Type)
	}
	return nil
}
