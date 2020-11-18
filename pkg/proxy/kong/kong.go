package kong

import (
	"fmt"
	"strings"
	"time"

	"github.com/caicloud/clientset/informers"
	"github.com/caicloud/clientset/kubernetes"
	lblisters "github.com/caicloud/clientset/listers/loadbalance/v1alpha2"
	lbapi "github.com/caicloud/clientset/pkg/apis/loadbalance/v1alpha2"
	controllerutil "github.com/caicloud/clientset/util/controller"
	"github.com/caicloud/clientset/util/syncqueue"
	"github.com/caicloud/loadbalancer-controller/pkg/api"
	"github.com/caicloud/loadbalancer-controller/pkg/config"
	"github.com/caicloud/loadbalancer-controller/pkg/plugin"
	lbutil "github.com/caicloud/loadbalancer-controller/pkg/util/lb"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	appslisters "k8s.io/client-go/listers/apps/v1"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog"
)

const (
	proxyName       = "kong"
	proxyNameSuffix = "-proxy-kong"
)

type kong struct {
	initialized bool

	proxyImage   string
	ingressImage string

	client kubernetes.Interface
	queue  *syncqueue.SyncQueue

	lbLister  lblisters.LoadBalancerLister
	dpLister  appslisters.DeploymentLister
	podLister corelisters.PodLister
}

func New() plugin.Interface {
	return &kong{}
}

func (k *kong) Init(cfg config.Configuration, factory informers.SharedInformerFactory) {
	if k.initialized {
		return
	}
	k.initialized = true

	klog.Info("Initialize the kong proxy")

	k.client = cfg.Client
	k.proxyImage = cfg.Proxies.Kong.ProxyImage
	k.ingressImage = cfg.Proxies.Kong.IngressImage

	if err := installKongCRDs(); err != nil {
		klog.Fatalf("Install kong crds failed: %v", err)
	}

	klog.Info("All kong crds installed.")

	dpInformer := factory.Apps().V1().Deployments()
	podInformer := factory.Core().V1().Pods()

	k.lbLister = factory.Loadbalance().V1alpha2().LoadBalancers().Lister()
	k.dpLister = dpInformer.Lister()
	k.podLister = podInformer.Lister()

	k.queue = syncqueue.NewPassthroughSyncQueue(&lbapi.LoadBalancer{}, k.syncLoadBalancer)
	dpInformer.Informer().AddEventHandler(lbutil.NewEventHandlerForDeployment(k.lbLister, k.dpLister, k.queue, k.deploymentFiltered))
	podInformer.Informer().AddEventHandler(lbutil.NewEventHandlerForSyncStatusWithPod(k.lbLister, k.podLister, k.queue, k.podFiltered))
}

func (k *kong) Run(stopCh <-chan struct{}) {
	if !k.initialized {
		klog.Fatal("Please initialize proxy before you run it")
	}
	defer utilruntime.HandleCrash()

	klog.Info("Starting kong proxy")
	defer func() {
		k.queue.ShutDown()
	}()

	k.queue.Run(1)
	<-stopCh
}

func (k *kong) OnSync(lb *lbapi.LoadBalancer) {
	klog.Infof("Syncing proxy, triggered by loadbalancer %v/%v", lb.Namespace, lb.Name)
	k.queue.Enqueue(lb)
}

func (k *kong) syncLoadBalancer(obj interface{}) error {
	lb, ok := obj.(*lbapi.LoadBalancer)
	if !ok {
		return fmt.Errorf("expect loadbalancer, got %v", obj)
	}

	key, _ := cache.DeletionHandlingMetaNamespaceKeyFunc(lb)

	startTime := time.Now()
	defer func() {
		klog.V(5).Infof("Finished syncing kong proxy for %v, usedTime %v", key, time.Since(startTime))
	}()

	nlb, err := k.lbLister.LoadBalancers(lb.Namespace).Get(lb.Name)
	if errors.IsNotFound(err) {
		klog.Warningf("LoadBalancer %v has been deleted, clean up proxy", key)
		return k.cleanup(lb)
	}

	if err != nil {
		utilruntime.HandleError(fmt.Errorf("unable to retrieve LoadBalancer %v from store: %v", key, err))
		return err
	}

	// fresh lb, original loadbalancer is gone
	if lb.UID != nlb.UID {
		return nil
	}

	if lb.DeletionTimestamp != nil {
		return nil
	}

	lb = nlb.DeepCopy()
	if lb.Spec.Proxy.Type != lbapi.ProxyTypeKong {
		klog.Infof("lb %v is not responsible for kong", lb.Name)
		return nil
	}

	ds, err := k.getDeploymentsForLoadBalancer(lb)
	if err != nil {
		return err
	}
	return k.sync(lb, ds)
}

func (k *kong) getDeploymentsForLoadBalancer(lb *lbapi.LoadBalancer) ([]*appsv1.Deployment, error) {
	// construct selector
	selector := k.selector(lb).AsSelector()

	// list all
	dList, err := k.dpLister.Deployments(lb.Namespace).List(selector)
	if err != nil {
		return nil, err
	}

	// If any adoptions are attempted, we should first recheck for deletion with
	// an uncached quorum read sometime after listing deployment (see kubernetes#42639).
	canAdoptFunc := controllerutil.RecheckDeletionTimestamp(func() (metav1.Object, error) {
		// fresh lb
		fresh, err := k.client.LoadbalanceV1alpha2().LoadBalancers(lb.Namespace).Get(lb.Name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}

		if fresh.UID != lb.UID {
			return nil, fmt.Errorf("original LoadBalancer %v/%v is gone: got uid %v, wanted %v", lb.Namespace, lb.Name, fresh.UID, lb.UID)
		}
		return fresh, nil
	})
	cm := controllerutil.NewDeploymentControllerRefManager(k.client, lb, selector, api.ControllerKind, canAdoptFunc)
	return cm.Claim(dList)
}

func (k *kong) sync(lb *lbapi.LoadBalancer, dps []*appsv1.Deployment) error {
	if err := installGlobalPlugins(lb.Name); err != nil {
		klog.Errorf("Install kong global plugins failed: %v", err)
		return err
	}

	desiredDeploy := k.generateDeployment(lb)

	deploymentName := ""
	// update
	var err error
	updated := false

	for _, dp := range dps {
		// two conditions will trigger controller to scale down deployment
		// 1. deployment does not have auto-generated prefix
		// 2. if there are more than one active controllers, there may be many valid deployments.
		//    But we only need one.
		if !strings.HasPrefix(dp.Name, lb.Name+proxyNameSuffix) || updated {
			if *dp.Spec.Replicas == 0 {
				continue
			}
			// scale unexpected deployment replicas to zero
			dpCopy := dp.DeepCopy()
			replica := int32(0)
			dpCopy.Spec.Replicas = &replica
			_, _ = k.client.AppsV1().Deployments(lb.Namespace).Update(dpCopy)
			continue
		}

		updated = true
		// do not change deployment if the loadbalancer is static
		if !lbutil.IsStatic(lb) {
			lbutil.InsertHelmAnnotation(desiredDeploy, dp.Namespace, dp.Name)
			merged, changed := lbutil.MergeDeployment(dp, desiredDeploy)
			deploymentName = merged.Name
			if changed {
				klog.Infof("Sync kong deployment %v for loadbalancer %v", dp.Name, lb.Name)
				if _, err = k.client.AppsV1().Deployments(lb.Namespace).Update(merged); err != nil {
					return err
				}
			}
		}
	}

	// len(dps) == 0 or no deployment's name match desired deployment
	if !updated {
		// create deployment
		klog.Infof("Create kong deployment %v for loadbalancer %v", desiredDeploy.Name, lb.Name)
		lbutil.InsertHelmAnnotation(desiredDeploy, desiredDeploy.Namespace, desiredDeploy.Name)
		if _, err = k.client.AppsV1().Deployments(lb.Namespace).Create(desiredDeploy); err != nil {
			return err
		}
		deploymentName = desiredDeploy.Name
	}

	if err := k.ensureService(lb); err != nil {
		return err
	}

	return k.syncStatus(lb, deploymentName)
}

// filter Deployment that controller does not care
func (k *kong) deploymentFiltered(obj *appsv1.Deployment) bool {
	return k.filteredByLabel(obj)
}

func (k *kong) podFiltered(obj *corev1.Pod) bool {
	return k.filteredByLabel(obj)
}

func (k *kong) filteredByLabel(obj metav1.ObjectMetaAccessor) bool {
	// obj.Labels
	selector := labels.Set{lbapi.LabelKeyProxy: proxyName}.AsSelector()
	match := selector.Matches(labels.Set(obj.GetObjectMeta().GetLabels()))

	return !match
}

func (k *kong) selector(lb *lbapi.LoadBalancer) labels.Set {
	return labels.Set{
		lbapi.LabelKeyCreatedBy: fmt.Sprintf(lbapi.LabelValueFormatCreateby, lb.Namespace, lb.Name),
		lbapi.LabelKeyProxy:     proxyName,
	}
}

// cleanup will delete all resources of this lb, such as daemonset and ingress.
func (k *kong) cleanup(lb *lbapi.LoadBalancer) error {
	dps, err := k.getDeploymentsForLoadBalancer(lb)
	if err != nil {
		return err
	}

	policy := metav1.DeletePropagationForeground
	gracePeriodSeconds := int64(30)

	for _, d := range dps {
		if err := k.client.AppsV1().Deployments(d.Namespace).Delete(d.Name, &metav1.DeleteOptions{
			GracePeriodSeconds: &gracePeriodSeconds,
			PropagationPolicy:  &policy,
		}); err != nil {
			klog.Errorf("Cleanup proxy failed: %v", err)
			return err
		}
	}

	// clean up ingress
	ingresses, err := k.client.ExtensionsV1beta1().Ingresses(metav1.NamespaceAll).List(metav1.ListOptions{
		LabelSelector: labels.FormatLabels(map[string]string{
			// created by IngressClass
			lbapi.LabelKeyCreatedBy: fmt.Sprintf(lbapi.LabelValueFormatCreateby, lb.Namespace, lb.Name),
		}),
	})
	if err != nil {
		klog.Errorf("List Ingress failed: %v", err)
		return err
	}

	for _, ingress := range ingresses.Items {
		if err = k.client.ExtensionsV1beta1().Ingresses(ingress.Namespace).Delete(ingress.Name, &metav1.DeleteOptions{}); err != nil {
			klog.Errorf("Cleanup Ingress failed: %v", err)
			return err
		}
	}
	return nil
}
