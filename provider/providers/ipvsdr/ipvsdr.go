package ipvsdr

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	log "github.com/zoumo/logdog"
	cli "gopkg.in/urfave/cli.v1"

	"github.com/caicloud/loadbalancer-controller/api"
	"github.com/caicloud/loadbalancer-controller/pkg/informers"
	listers "github.com/caicloud/loadbalancer-controller/pkg/listers/caicloud/v1beta1"
	"github.com/caicloud/loadbalancer-controller/pkg/tprclient"
	controllerutil "github.com/caicloud/loadbalancer-controller/pkg/util/controller"
	lbutil "github.com/caicloud/loadbalancer-controller/pkg/util/lb"
	"github.com/caicloud/loadbalancer-controller/pkg/util/validation"
	"github.com/caicloud/loadbalancer-controller/provider"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	extensionslisters "k8s.io/client-go/listers/extensions/v1beta1"
	"k8s.io/client-go/pkg/api/v1"
	extensions "k8s.io/client-go/pkg/apis/extensions/v1beta1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/kubernetes/pkg/controller"
)

const defaultImage = "cargo.caicloud.io/caicloud/ingress-nginx:v0.1.0"

// controllerKind contains the schema.GroupVersionKind for this controller type.
var controllerKind = api.SchemeGroupVersion.WithKind(api.LoadBalancerKind)

func init() {
	provider.RegisterPlugin("ipvsdr", NewIpvsdr())
}

var _ provider.Plugin = &ipvsdr{}

type ipvsdr struct {
	initialized bool

	image string

	client    kubernetes.Interface
	tprclient tprclient.Interface

	helper       *controllerutil.Helper
	eventHandler *lbutil.EventHandlerForDeployment

	lbLister       listers.LoadBalancerLister
	dLister        extensionslisters.DeploymentLister
	lbListerSynced cache.InformerSynced
	dListerSynced  cache.InformerSynced

	queue workqueue.RateLimitingInterface
}

// NewIpvsdr creates a new ipvsdr provider plugin
func NewIpvsdr() provider.Plugin {
	return &ipvsdr{
		image: defaultImage,
	}
}

func (f *ipvsdr) AddFlags(app *cli.App) {

	flags := []cli.Flag{
		cli.StringFlag{
			Name:        "provider-ipvsdr",
			Usage:       "ipvsdr provider image",
			EnvVar:      "PROVIDER_IPVS_DR",
			Value:       defaultImage,
			Destination: &f.image,
		},
	}
	app.Flags = append(app.Flags, flags...)
}

func (f *ipvsdr) Init(sif informers.SharedInformerFactory) {
	if f.initialized {
		return
	}

	log.Info("Initialize the ipvsdr provider")

	f.client = sif.Client()
	f.tprclient = sif.TPRClient()

	lbInformer := sif.Caicloud().V1beta1().LoadBalancer()
	dInformer := sif.Extensions().V1beta1().Deployments()

	f.lbLister = lbInformer.Lister()
	f.dLister = dInformer.Lister()
	f.lbListerSynced = lbInformer.Informer().HasSynced
	f.dListerSynced = dInformer.Informer().HasSynced

	f.queue = workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "provider-ipvsdr")
	f.helper = controllerutil.NewHelperForObj(&api.LoadBalancer{}, f.queue, f.syncLoadBalancer)
	f.eventHandler = lbutil.NewEventHandlerForDeployment(lbInformer, dInformer, f.helper, f.filtered)

	dInformer.Informer().AddEventHandler(f.eventHandler)

	f.initialized = true
}

func (f *ipvsdr) Run(stopCh <-chan struct{}) {

	workers := 1

	if !f.initialized {
		log.Panic("Please initialize provider before you run it")
		return
	}

	defer utilruntime.HandleCrash()
	defer f.queue.ShutDown()

	log.Info("Starting ipvsdr provider", log.Fields{"workers": workers, "image": f.image})
	defer log.Info("Shutting down ipvsdr provider")

	if !cache.WaitForCacheSync(stopCh, f.lbListerSynced, f.dListerSynced) {
		log.Error("Wait for cache sync timeout")
		return
	}

	for i := 0; i < workers; i++ {
		go wait.Until(f.helper.Worker, time.Second, stopCh)
	}

	<-stopCh

}

func (f *ipvsdr) filtered(obj *extensions.Deployment) bool {
	// obj.Labels
	selector := labels.Set{api.LabelKeyProvider: "ipvsdr"}.AsSelector()
	match := selector.Matches(labels.Set(obj.Labels))

	return !match
}

func (f *ipvsdr) OnSync(lb *api.LoadBalancer) {
	if lb.Type != api.LoadBalancerTypeExternal && lb.Providers.Ipvsdr != nil {
		// It is not my responsible
		return
	}
	log.Info("Syncing providers, triggered by lb controller", log.Fields{"lb": lb.Name, "namespace": lb.Namespace})
	f.helper.Enqueue(lb)
}

func (f *ipvsdr) syncLoadBalancer(obj interface{}) error {
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
		log.Info("Finished syncing ipvsdr provider", log.Fields{"lb": key, "usedTime": time.Now().Sub(startTime)})
	}()

	nlb, err := f.lbLister.LoadBalancers(lb.Namespace).Get(lb.Name)
	if errors.IsNotFound(err) {
		log.Warn("LoadBalancer has been deleted, clean up provider", log.Fields{"lb": key})

		return f.cleanup(lb)
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

	ds, err := f.getDeploymentsForLoadBalancer(lb)
	if err != nil {
		return err
	}

	if lb.DeletionTimestamp != nil {
		// TODO sync status only
		return nil
	}

	return f.sync(lb, ds)
}

func (f *ipvsdr) getDeploymentsForLoadBalancer(lb *api.LoadBalancer) ([]*extensions.Deployment, error) {

	// construct selector
	selector := labels.Set{
		api.LabelKeyCreateby: fmt.Sprintf(api.LabelValueFormatCreateby, lb.Namespace, lb.Name),
		api.LabelKeyProvider: "ipvsdr",
	}.AsSelector()

	// list all
	dList, err := f.dLister.Deployments(lb.Namespace).List(selector)
	if err != nil {
		return nil, err
	}

	// If any adoptions are attempted, we should first recheck for deletion with
	// an uncached quorum read sometime after listing deployment (see kubernetes#42639).
	canAdoptFunc := controller.RecheckDeletionTimestamp(func() (metav1.Object, error) {
		// fresh lb
		fresh, err := f.tprclient.CaicloudV1beta1().LoadBalancers(lb.Namespace).Get(lb.Name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}

		if fresh.UID != lb.UID {
			return nil, fmt.Errorf("original LoadBalancer %v/%v is gone: got uid %v, wanted %v", lb.Namespace, lb.Name, fresh.UID, lb.UID)
		}
		return fresh, nil
	})

	cm := controllerutil.NewDeploymentControllerRefManager(f.client, lb, selector, controllerKind, canAdoptFunc)
	return cm.Claim(dList)
}

// sync generate desired deployment from lb and compare it with existing deployment
func (f *ipvsdr) sync(lb *api.LoadBalancer, dps []*extensions.Deployment) error {
	desiredDeploy := f.generateDeployment(lb)

	// update
	updated := false
	var err error
	for _, dp := range dps {
		// auto generate deployment has this prefix
		if !strings.HasPrefix(dp.Name, lb.Name+"-provider-ipvsdr") {
			if *dp.Spec.Replicas == 0 {
				continue
			}
			// scale unexpected deployment replicas to zero
			log.Info("Scale unexpected provider replicas to zero", log.Fields{"d.name": dp.Name, "lb.name": lb.Name})
			copy, _ := lbutil.DeploymentDeepCopy(dp)
			replica := int32(0)
			copy.Spec.Replicas = &replica
			f.client.ExtensionsV1beta1().Deployments(lb.Namespace).Update(copy)
			continue
		}
		updated = true
		copyDp, changed, newErr := f.ensureDeployment(desiredDeploy, dp)
		if newErr != nil {
			err = newErr
			continue
		}
		if changed {
			log.Info("Sync deployment for lb", log.Fields{"d.name": dp.Name, "lb.name": lb.Name})
			_, err = f.client.ExtensionsV1beta1().Deployments(lb.Namespace).Update(copyDp)
		}
	}

	// len(dps) == 0 or no deployment's name match desired deployment
	if !updated {
		// create deployment
		log.Info("Create deployment for lb", log.Fields{"d.name": desiredDeploy.Name, "lb.name": lb.Name})
		_, err := f.client.ExtensionsV1beta1().Deployments(lb.Namespace).Create(desiredDeploy)
		return err
	}

	return err
}

func (f *ipvsdr) ensureDeployment(desiredDeploy, oldDeploy *extensions.Deployment) (*extensions.Deployment, bool, error) {
	copyDp, err := lbutil.DeploymentDeepCopy(oldDeploy)
	if err != nil {
		return nil, false, err
	}

	// ensure labels
	for k, v := range desiredDeploy.Labels {
		copyDp.Labels[k] = v
	}

	// ensure replicas
	copyDp.Spec.Replicas = desiredDeploy.Spec.Replicas

	labelChanged := !reflect.DeepEqual(oldDeploy, copyDp)
	replicasChanged := *copyDp.Spec.Replicas != *oldDeploy.Spec.Replicas

	changed := labelChanged || replicasChanged
	if changed {
		log.Info("Abount to change deployment", log.Fields{
			"dp.name":         copyDp.Name,
			"labelChanged":    labelChanged,
			"replicasChanged": replicasChanged,
		})
	}

	return copyDp, changed, nil
}

// cleanup deployment and other resource controlled by ipvsdr provider
func (f *ipvsdr) cleanup(lb *api.LoadBalancer) error {

	ds, err := f.getDeploymentsForLoadBalancer(lb)
	if err != nil {
		return err
	}

	for _, d := range ds {
		f.client.ExtensionsV1beta1().Deployments(d.Namespace).Delete(d.Name, &metav1.DeleteOptions{})
	}

	return nil
}

func (f *ipvsdr) generateDeployment(lb *api.LoadBalancer) *extensions.Deployment {
	terminationGracePeriodSeconds := int64(30)
	hostNetwork := true
	replicas := lbutil.CalculateReplicas(lb)
	privileged := true

	labels := map[string]string{
		api.LabelKeyCreateby: fmt.Sprintf(api.LabelValueFormatCreateby, lb.Namespace, lb.Name),
		api.LabelKeyProvider: "ipvsdr",
	}

	// run in this node
	nodeAffinity := &v1.NodeAffinity{
		RequiredDuringSchedulingIgnoredDuringExecution: &v1.NodeSelector{
			NodeSelectorTerms: []v1.NodeSelectorTerm{
				{
					MatchExpressions: []v1.NodeSelectorRequirement{
						{
							Key:      fmt.Sprintf(api.UniqueLabelKeyFormat, lb.Namespace, lb.Name),
							Operator: v1.NodeSelectorOpIn,
							Values:   []string{"true"},
						},
					},
				},
			},
		},
	}

	// do not run with this pod
	podAffinity := &v1.PodAntiAffinity{
		RequiredDuringSchedulingIgnoredDuringExecution: []v1.PodAffinityTerm{
			{
				LabelSelector: &metav1.LabelSelector{
					MatchLabels: labels,
				},
				TopologyKey: metav1.LabelHostname,
			},
		},
	}

	// privileged := true

	t := true

	deploy := &extensions.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:   lb.Name + "-provider-ipvsdr-" + lbutil.RandStringBytesRmndr(5),
			Labels: labels,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion:         controllerKind.GroupVersion().String(),
					Kind:               controllerKind.Kind,
					Name:               lb.Name,
					UID:                lb.UID,
					Controller:         &t,
					BlockOwnerDeletion: &t,
				},
			},
		},
		Spec: extensions.DeploymentSpec{
			Replicas: &replicas,
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: v1.PodSpec{
					// host network ?
					HostNetwork: hostNetwork,
					// TODO
					TerminationGracePeriodSeconds: &terminationGracePeriodSeconds,
					Affinity: &v1.Affinity{
						// decide running on which node
						NodeAffinity: nodeAffinity,
						// don't co-locate pods of this deployment in same node
						PodAntiAffinity: podAffinity,
					},
					// tolerate taints
					Tolerations: []v1.Toleration{
						{
							Key:      api.TaintKey,
							Operator: v1.TolerationOpExists,
						},
					},
					Containers: []v1.Container{
						{
							Name:            "ipvsdr",
							Image:           f.image,
							ImagePullPolicy: v1.PullAlways,
							Resources: v1.ResourceRequirements{
								Limits: v1.ResourceList{
									v1.ResourceCPU:    resource.MustParse("1000m"),
									v1.ResourceMemory: resource.MustParse("500Mi"),
								},
							},
							SecurityContext: &v1.SecurityContext{
								Privileged: &privileged,
							},
							Ports: []v1.ContainerPort{
								{
									ContainerPort: 80,
								},
								{
									ContainerPort: 443,
								},
							},
							// TODO
							Args: []string{
								// watch on which lb
								"--loadbalancer=" + lb.Namespace + "/" + lb.Name,
							},
							// ReadinessProbe: &v1.Probe{
							// 	Handler: v1.Handler{
							// 		HTTPGet: &v1.HTTPGetAction{
							// 			Path:   "/ingress-controller-healthz",
							// 			Port:   intstr.FromInt(10254),
							// 			Scheme: v1.URISchemeHTTP,
							// 		},
							// 	},
							// },
							// LivenessProbe: &v1.Probe{
							// 	Handler: v1.Handler{
							// 		HTTPGet: &v1.HTTPGetAction{
							// 			Path:   "/ingress-controller-healthz",
							// 			Port:   intstr.FromInt(10254),
							// 			Scheme: v1.URISchemeHTTP,
							// 		},
							// 	},
							// },
						},
					},
				},
			},
		},
	}

	return deploy
}
