package external

import (
	"fmt"

	"github.com/caicloud/clientset/informers"
	"github.com/caicloud/clientset/kubernetes"
	lblisters "github.com/caicloud/clientset/listers/loadbalance/v1alpha2"
	lbapi "github.com/caicloud/clientset/pkg/apis/loadbalance/v1alpha2"
	"github.com/caicloud/clientset/util/syncqueue"
	"github.com/caicloud/loadbalancer-controller/pkg/config"
	"github.com/caicloud/loadbalancer-controller/pkg/provider"
	lbutil "github.com/caicloud/loadbalancer-controller/pkg/util/lb"
	"github.com/caicloud/loadbalancer-controller/pkg/util/validation"
	log "github.com/zoumo/logdog"

	"k8s.io/apimachinery/pkg/api/errors"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
)

const (
	providerName = "external"
)

func init() {
	provider.RegisterPlugin(providerName, NewExternal())
}

var _ provider.Plugin = &external{}

type external struct {
	initialized bool

	client kubernetes.Interface
	queue  *syncqueue.SyncQueue

	lbLister lblisters.LoadBalancerLister
}

// NewExternal creates a new external provider plugin
func NewExternal() provider.Plugin {
	return &external{}
}

func (f *external) Init(cfg config.Configuration, sif informers.SharedInformerFactory) {
	if f.initialized {
		return
	}
	f.initialized = true
	log.Info("Initialize the external provider")

	// set config
	f.client = cfg.Client
	// initialize controller
	lbInformer := sif.Loadbalance().V1alpha2().LoadBalancers()
	f.lbLister = lbInformer.Lister()
	f.queue = syncqueue.NewPassthroughSyncQueue(&lbapi.LoadBalancer{}, f.syncLoadBalancer)

}

func (f *external) Run(stopCh <-chan struct{}) {

	workers := 1

	if !f.initialized {
		log.Panic("Please initialize provider before you run it")
		return
	}

	defer utilruntime.HandleCrash()

	log.Info("Starting external provider", log.Fields{"workers": workers})
	defer log.Info("Shutting down external provider")

	// lb controller has waited all the informer synced
	// there is no need to wait again here

	defer func() {
		log.Info("Shutting down external provider")
		f.queue.ShutDown()
	}()

	f.queue.Run(workers)

	<-stopCh
}

func (f *external) OnSync(lb *lbapi.LoadBalancer) {
	if lb.Spec.Providers.External == nil {
		// It is not my responsible
		return
	}
	log.Info("Syncing providers, triggered by lb controller", log.Fields{"lb": lb.Name, "namespace": lb.Namespace})
	f.queue.Enqueue(lb)
}

func (f *external) syncLoadBalancer(obj interface{}) error {
	lb, ok := obj.(*lbapi.LoadBalancer)
	if !ok {
		return fmt.Errorf("expect loadbalancer, got %v", obj)
	}

	// Validate loadbalancer scheme
	if err := validation.ValidateLoadBalancer(lb); err != nil {
		log.Debug("invalid loadbalancer scheme", log.Fields{"err": err})
		return err
	}

	key, _ := cache.DeletionHandlingMetaNamespaceKeyFunc(lb)

	nlb, err := f.lbLister.LoadBalancers(lb.Namespace).Get(lb.Name)
	if errors.IsNotFound(err) {
		return nil
	}
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("Unable to retrieve LoadBalancer %v from store: %v", key, err))
		return err
	}

	// fresh lb
	if lb.UID != nlb.UID {
		return nil
	}

	lb = nlb

	if lb.DeletionTimestamp != nil {
		// TODO sync status only
		return nil
	}

	// sync status
	providerStatus := lbapi.ExpternalProviderStatus{
		VIP: lb.Spec.Providers.External.VIP,
	}
	externalstatus := lb.Status.ProvidersStatuses.External
	// check whether the statuses are equal
	if externalstatus == nil || !lbutil.ExternalProviderStatusEqual(*externalstatus, providerStatus) {
		_, err := lbutil.UpdateLBWithRetries(
			f.client.LoadbalanceV1alpha2().LoadBalancers(lb.Namespace),
			f.lbLister,
			lb.Namespace,
			lb.Name,
			func(lb *lbapi.LoadBalancer) error {
				lb.Status.ProvidersStatuses.External = &providerStatus
				return nil
			},
		)

		if err != nil {
			log.Error("Update loadbalancer status error", log.Fields{"err": err})
			return err
		}

	}

	return nil
}
