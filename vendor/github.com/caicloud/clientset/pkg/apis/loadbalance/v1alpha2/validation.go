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
		if net.ParseIP(ipvsdr.VIP) == nil {
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
	if spec.External != nil {
		external := spec.External
		if net.ParseIP(external.VIP) == nil {
			return fmt.Errorf("external: vip is invalid")
		}
	}
	if spec.Azure != nil {
		azure := spec.Azure
		if len(azure.Location) == 0 {
			return fmt.Errorf("azure: location cant't be empty")
		}
		if len(azure.ResourceGroupName) == 0 {
			return fmt.Errorf("azure: group name cant't be empty")
		}
		if len(azure.ClusterID) == 0 {
			return fmt.Errorf("azure: cluster id cant't be empty")
		}
		if azure.SKU != AzureStandardSKU && azure.SKU != AzureBasicSKU {
			return fmt.Errorf("azure: sku %v is invalid", azure.SKU)
		}
		if azure.IPAddressProperties.Public == nil && azure.IPAddressProperties.Private == nil {
			return fmt.Errorf("azure: private and public ip address can't be nil at the same time")
		}
		if azure.IPAddressProperties.Public != nil && azure.IPAddressProperties.Private != nil {
			return fmt.Errorf("azure: private and public ip address can't have value at the same time")
		}
		if azure.IPAddressProperties.Private != nil {
			private := azure.IPAddressProperties.Private
			if len(private.SubnetID) == 0 {
				return fmt.Errorf("azure: subnet id cant't be empty when use private network")
			}
			switch private.IPAllocationMethod {
			case AzureStaticIPAllocationMethod:
				if private.PrivateIPAddress == nil {
					return fmt.Errorf("azure: private ip address can't be nil when allocation method is static")
				}
				if net.ParseIP(*private.PrivateIPAddress) == nil {
					return fmt.Errorf("azure: private ip is invalid")
				}
			//TODO don't support
			case AzureDynamicIPAllocationMethod:
			default:
				return fmt.Errorf("azure: private allocation method %v is invalid", private.IPAllocationMethod)
			}
		}
		if azure.IPAddressProperties.Public != nil {
			public := azure.IPAddressProperties.Public
			switch public.IPAllocationMethod {
			case AzureStaticIPAllocationMethod:
				if public.PublicIPAddressID == nil {
					return fmt.Errorf("azure: public ip address can't be nil when allocation method is static")
				}
			//TODO don't support
			case AzureDynamicIPAllocationMethod:
			default:
				return fmt.Errorf("azure: public allocation method %v is invalid", public.IPAllocationMethod)
			}
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
