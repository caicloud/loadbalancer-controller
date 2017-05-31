package controller

import (
	"time"

	lbapi "github.com/caicloud/loadbalancer-controller/api"
	log "github.com/zoumo/logdog"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	apiv1 "k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"
	"k8s.io/client-go/rest"
)

// EnsureTPR initialize third party resource if it does not exist
func EnsureTPR(clientset kubernetes.Interface) error {
	tpr := &v1beta1.ThirdPartyResource{
		ObjectMeta: metav1.ObjectMeta{
			// this kild of objects will be LoadBalancer
			// More info: https://kubernetes.io/docs/tasks/access-kubernetes-api/extend-api-third-party-resource/
			Name: lbapi.LoadBalancerTPRName + "." + lbapi.GroupName,
		},
		Versions: []v1beta1.APIVersion{
			{Name: lbapi.Version},
		},
		Description: "A specification of loadbalancer to provider load balancing for ingress",
	}

	_, err := clientset.ExtensionsV1beta1().ThirdPartyResources().Create(tpr)

	if err != nil && apierrors.IsAlreadyExists(err) {
		log.Info("Skip the creation for ThirdPartyResource LoadBalancer because it has already been created")
		return nil
	}

	return err
}

func WaitForLoadBalancerResource(client *rest.RESTClient) error {
	return wait.Poll(100*time.Millisecond, 60*time.Second, func() (bool, error) {
		_, err := client.Get().Namespace(apiv1.NamespaceDefault).Resource(lbapi.LoadBalancerPlural).DoRaw()
		if err == nil {
			return true, nil
		}
		if apierrors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	})
}
