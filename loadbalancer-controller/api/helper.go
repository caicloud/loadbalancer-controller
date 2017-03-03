package api


import (
	"k8s.io/client-go/1.5/pkg/runtime"
	"k8s.io/client-go/1.5/pkg/util/json"
)

// ToLoadbalancerClaim converts  a Unstructured to LoadBalancerClaim
func ToLoadbalancerClaim(unstructed *runtime.Unstructured) (*LoadBalancerClaim, error) {
	data, err := unstructed.MarshalJSON()
	if err != nil {
		return nil, err
	}
	claim := &LoadBalancerClaim{}
	if err := json.Unmarshal(data, claim); err != nil {
		return nil, err
	}
	return claim, nil
}

// ToUnstructured converts  a LoadBalancerClaim to ToUnstructured
func (claim LoadBalancerClaim) ToUnstructured() (*runtime.Unstructured, error) {
	data, err := json.Marshal(claim)
	if err != nil {
		return nil, err
	}
	unstructed := &runtime.Unstructured{}
	if err := unstructed.UnmarshalJSON(data); err != nil {
		return nil, err
	}
	return unstructed, nil
}

// ToUnstructured converts  a LoadBalancer to ToUnstructured
func (lb LoadBalancer) ToUnstructured() (*runtime.Unstructured, error) {
	data, err := json.Marshal(lb)
	if err != nil {
		return nil, err
	}
	unstructed := &runtime.Unstructured{}
	if err := unstructed.UnmarshalJSON(data); err != nil {
		return nil, err
	}
	return unstructed, nil
}