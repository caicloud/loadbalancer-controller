package controller

import (
	"fmt"
	"reflect"
	"time"

	"github.com/caicloud/loadbalancer-controller/api"
	"github.com/caicloud/loadbalancer-controller/pkg/tprclient"
	controllerutil "github.com/caicloud/loadbalancer-controller/pkg/util/controller"
	"github.com/caicloud/loadbalancer-controller/pkg/util/validation"

	lbapi "github.com/caicloud/loadbalancer-controller/api"
	"github.com/caicloud/loadbalancer-controller/pkg/informers"
	listers "github.com/caicloud/loadbalancer-controller/pkg/listers/caicloud/v1beta1"
	"github.com/caicloud/loadbalancer-controller/pkg/util/taints"
	"github.com/caicloud/loadbalancer-controller/provider"
	"github.com/caicloud/loadbalancer-controller/proxy"
	log "github.com/zoumo/logdog"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	corelisters "k8s.io/client-go/listers/core/v1"
	apiv1 "k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

const (
	maxRetries = 5
)

// LoadBalancerController is responsible for synchronizing LoadBalancer objects stored
// in the system with actual running proxies and providers.
type LoadBalancerController struct {
	kubeClient kubernetes.Interface
	tprClient  tprclient.Interface

	factory    informers.SharedInformerFactory
	lbLister   listers.LoadBalancerLister
	nodeLister corelisters.NodeLister

	queue  workqueue.RateLimitingInterface
	helper *controllerutil.Helper
}

// NewLoadBalancerController creates a new LoadBalancerController.
func NewLoadBalancerController(client kubernetes.Interface, tprClient tprclient.Interface) *LoadBalancerController {
	// TODO register metrics

	lbc := &LoadBalancerController{
		kubeClient: client,
		tprClient:  tprClient,
		factory:    informers.NewSharedInformerFactory(client, tprClient, 0),
		queue:      workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "loadbalancer"),
	}

	// setup lb controller helper
	lbc.helper = controllerutil.NewHelperForObj(&lbapi.LoadBalancer{}, lbc.queue, lbc.syncLoadBalancer)

	// setup informer
	lbinformer := lbc.factory.Caicloud().V1beta1().LoadBalancer()
	lbinformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    lbc.addLoadBalancer,
		UpdateFunc: lbc.updateLoadBalancer,
		DeleteFunc: lbc.deleteLoadBalancer,
	})

	lbc.lbLister = lbinformer.Lister()
	lbc.nodeLister = lbc.factory.Core().V1().Nodes().Lister()

	// setup proxies
	proxy.Init(lbc.factory)
	// setup providers
	provider.Init(lbc.factory)

	return lbc
}

// Run begins watching and syncing.
func (lbc *LoadBalancerController) Run(workers int, stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	defer lbc.queue.ShutDown()

	log.Info("Startting loadbalancer controller")
	defer log.Info("Shutting down loadbalancer controller")

	// ensure loadbalancer tpr initialized
	if err := lbc.ensureResource(); err != nil {
		log.Error("Ensure loadbalancer resource error", log.Fields{"err": err})
		return
	}

	// start shared informer
	lbc.factory.Start(stopCh)

	// wait cache synced
	log.Info("Wait for all caches syncing")
	synced := lbc.factory.WaitForCacheSync(stopCh)
	for tpy, sync := range synced {
		if !sync {
			log.Error("Wait for cache sync timeout", log.Fields{"type": tpy})
			return
		}
	}
	log.Info("All caches have synced, Running LoadBalancer Controller ...", log.Fields{"worker": workers})

	// start loadbalancer worker
	for i := 0; i < workers; i++ {
		go wait.Until(lbc.helper.Worker, time.Second, stopCh)
	}

	// run proxy
	proxy.Run(stopCh)
	// run providers
	provider.Run(stopCh)

	<-stopCh
}

// ensure loadbalancer tpr initialized
func (lbc *LoadBalancerController) ensureResource() error {
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

	_, err := lbc.kubeClient.ExtensionsV1beta1().ThirdPartyResources().Create(tpr)

	if err != nil && errors.IsAlreadyExists(err) {
		log.Info("Skip the creation for ThirdPartyResource LoadBalancer because it has already been created")
		return nil
	}

	return err
}

// syncLoadBalancer will sync the loadbalancer with the given key.
// This function is not meant to be invoked concurrently with the same key.
func (lbc *LoadBalancerController) syncLoadBalancer(obj interface{}) error {
	lb, ok := obj.(*api.LoadBalancer)
	if !ok {
		return fmt.Errorf("expect loadbalancer, got %v", obj)
	}

	// Validate loadbalancer scheme
	if err := validation.ValidateLoadBalancer(lb); err != nil {
		log.Debug("invalid loadbalancer scheme", log.Fields{"err": err})
		return err
	}

	key, _ := controllerutil.KeyFunc(lb)

	startTime := time.Now()
	defer func() {
		log.Info("Finished syncing loadbalancer", log.Fields{"key": key, "usedTime": time.Now().Sub(startTime)})
	}()

	nlb, err := lbc.lbLister.LoadBalancers(lb.Namespace).Get(lb.Name)
	if errors.IsNotFound(err) {
		log.Warn("LoadBalancer has been deleted", log.Fields{"lb": key})
		// deleted
		return lbc.sync(lb, true)
	}
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("Unable to retrieve LoadBalancer %v from store: %v", key, err))
		return err
	}

	// fresh lb
	if lb.UID != nlb.UID {
		//  original loadbalancer is gone
		return nil
	}
	lb = nlb

	return lbc.sync(lb, false)
}

func (lbc *LoadBalancerController) sync(lb *lbapi.LoadBalancer, deleted bool) error {

	// sync proxy
	proxy.OnSync(lb)
	// sync provider
	provider.OnSync(lb)

	// sync nodes
	if deleted {
		nlb, err := lbc.clone(lb)
		if err != nil {
			return err
		}

		replicas := int32(0)
		nlb.Replicas = &replicas
		nlb.Nodes = &lbapi.NodesSpec{
			Names: []string{},
		}
		lb = nlb
	}

	// sync nodes
	err := lbc.syncNodes(lb)

	return err
}

func (lbc *LoadBalancerController) syncNodes(lb *lbapi.LoadBalancer) error {
	// varify desired nodes
	desiredNodes, err := lbc.getVerifiedNodes(lb)
	if err != nil {
		log.Error("varify nodes error", log.Fields{"err": err})
		return err
	}

	oldNodes, err := lbc.getNodesForLoadBalancer(lb)
	if err != nil {
		log.Error("list node error")
		return err
	}
	// compute diff
	nodesToDelete := lbc.nodesDiff(oldNodes, desiredNodes.Nodes)
	lbc.doLabelAndTaints(nodesToDelete, desiredNodes)
	return nil
}

func (lbc *LoadBalancerController) getNodesForLoadBalancer(lb *lbapi.LoadBalancer) ([]*apiv1.Node, error) {
	// list old nodes
	labelkey := fmt.Sprintf(lbapi.UniqueLabelKeyFormat, lb.Namespace, lb.Name)
	selector := labels.Set{labelkey: "true"}.AsSelector()
	return lbc.nodeLister.List(selector)
}

func (lbc *LoadBalancerController) nodesDiff(oldNodes, desiredNodes []*apiv1.Node) []*apiv1.Node {

	if len(desiredNodes) == 0 {
		return oldNodes
	}

	nodesToDelete := make([]*apiv1.Node, 0)

NEXT:
	for _, oldNode := range oldNodes {
		for _, desiredNode := range desiredNodes {
			if oldNode.Name == desiredNode.Name {
				continue NEXT
			}
		}
		nodesToDelete = append(nodesToDelete, oldNode)
	}

	return nodesToDelete
}

func (lbc *LoadBalancerController) addLoadBalancer(obj interface{}) {
	lb := obj.(*lbapi.LoadBalancer)
	log.Info("Adding LoadBalancer", log.Fields{"name": lb.Name})
	lbc.helper.Enqueue(lb)
}

func (lbc *LoadBalancerController) updateLoadBalancer(oldObj, curObj interface{}) {
	old := oldObj.(*lbapi.LoadBalancer)
	cur := curObj.(*lbapi.LoadBalancer)

	if old.ResourceVersion == cur.ResourceVersion {
		// Periodic resync will send update events for all known LoadBalancer.
		// Two different versions of the same LoadBalancer will always have different RVs.
		return
	}

	// can not change loadbalancer type from internal to external
	if old.Type == lbapi.LoadBalancerTypeInternal && cur.Type == lbapi.LoadBalancerTypeExternal {
		log.Warn("Forbidden to change the type of loadblancer, revert it", log.Fields{"from": old.Type, "to": cur.Type})
		revert, err := lbc.clone(cur)
		if err != nil {
			return
		}
		revert.LoadBalancerSpec.Type = old.Type
		_, err = lbc.tprClient.CaicloudV1beta1().LoadBalancers(cur.Namespace).Update(revert)
		if err != nil {
			log.Error("revert loadbalancer type error", log.Fields{"err": err})
		}

		return
	}

	log.Info("Updating LoadBalancer", log.Fields{"name": old.Name})
	lbc.helper.Enqueue(cur)
}

func (lbc *LoadBalancerController) deleteLoadBalancer(obj interface{}) {
	lb, ok := obj.(*lbapi.LoadBalancer)

	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("Couldn't get object from tombstone %#v", obj))
			return
		}
		lb, ok = tombstone.Obj.(*lbapi.LoadBalancer)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("Tombstone contained object that is not a LoadBalancer %#v", obj))
			return
		}
	}

	log.Info("Deleting LoadBalancer", log.Fields{"name": lb.Name})

	lbc.helper.Enqueue(lb)
}

func (lbc *LoadBalancerController) clone(lb *lbapi.LoadBalancer) (*lbapi.LoadBalancer, error) {
	lbi, err := scheme.Scheme.DeepCopy(lb)
	if err != nil {
		log.Error("Unable to deepcopy loadbalancer", log.Fields{"lb.name": lb.Name, "err": err})
		return nil, err
	}

	nlb, ok := lbi.(*lbapi.LoadBalancer)
	if !ok {
		nerr := fmt.Errorf("expected loadbalancer, got %#v", lbi)
		log.Error(nerr)
		return nil, err
	}
	return nlb, nil
}

// doLabelAndTaints delete label and taints in nodesToDelete
// add label and taints in nodes
func (lbc *LoadBalancerController) doLabelAndTaints(nodesToDelete []*apiv1.Node, desiredNodes *VerifiedNodes) {
	// delete labels and taints from old nodes
	for _, node := range nodesToDelete {
		copy, _ := scheme.Scheme.DeepCopy(node)
		copyNode := copy.(*apiv1.Node)

		// change labels
		for key := range desiredNodes.Labels {
			delete(copyNode.Labels, key)
		}

		// change taints
		// maybe taints are not found, reorganize will return error but it doesn't not matter
		// taints will not be changed
		_, newTaints, _ := taints.ReorganizeTaints(copyNode, false, nil, []apiv1.Taint{
			{Key: lbapi.TaintKey},
		})
		copyNode.Spec.Taints = newTaints

		labelChanged := !reflect.DeepEqual(node.Labels, copyNode.Labels)
		taintChanged := !reflect.DeepEqual(node.Spec.Taints, copyNode.Spec.Taints)
		if labelChanged || taintChanged {
			log.Debug("Delete labels and taints from old nodes", log.Fields{
				"node": node.Name,
			})
			_, err := lbc.kubeClient.CoreV1().Nodes().Update(copyNode)
			if err != nil {
				log.Errorf("update node err: %v", err)
			}
		}

	}

	// ensure labels and taints in cur nodes
	for _, node := range desiredNodes.Nodes {
		copy, _ := scheme.Scheme.DeepCopy(node)
		copyNode := copy.(*apiv1.Node)

		// change labels
		for k, v := range desiredNodes.Labels {
			copyNode.Labels[k] = v
		}

		// override taint, add or delete
		_, newTaints, _ := taints.ReorganizeTaints(copyNode, true, desiredNodes.TaintsToAdd, desiredNodes.TaintsToDelete)
		// If you don't judge， it maybe change from nil to []Taint{}
		// do not change taints when orgin and new length are 0
		if !(len(copyNode.Spec.Taints) == 0 && len(newTaints) == 0) {
			copyNode.Spec.Taints = newTaints
		}

		labelChanged := !reflect.DeepEqual(node.Labels, copyNode.Labels)
		taintChanged := !reflect.DeepEqual(node.Spec.Taints, copyNode.Spec.Taints)
		if labelChanged || taintChanged {
			log.Debug("Ensure labels and taints for requested nodes", log.Fields{
				"node":         node.Name,
				"labels":       node.Labels,
				"taints":       newTaints,
				"labelChanged": labelChanged,
				"taintChanged": taintChanged,
			})
			lbc.kubeClient.CoreV1().Nodes().Update(copyNode)
		}
	}

}
