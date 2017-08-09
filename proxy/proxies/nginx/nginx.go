/*
Copyright 2017 Caicloud authors. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package nginx

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/caicloud/loadbalancer-controller/config"
	netv1alpha1 "github.com/caicloud/loadbalancer-controller/pkg/apis/networking/v1alpha1"
	"github.com/caicloud/loadbalancer-controller/pkg/informers"
	netlisters "github.com/caicloud/loadbalancer-controller/pkg/listers/networking/v1alpha1"
	"github.com/caicloud/loadbalancer-controller/pkg/toleration"
	"github.com/caicloud/loadbalancer-controller/pkg/tprclient"
	controllerutil "github.com/caicloud/loadbalancer-controller/pkg/util/controller"
	lbutil "github.com/caicloud/loadbalancer-controller/pkg/util/lb"
	"github.com/caicloud/loadbalancer-controller/pkg/util/validation"
	"github.com/caicloud/loadbalancer-controller/proxy"
	log "github.com/zoumo/logdog"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/intstr"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	corelisters "k8s.io/client-go/listers/core/v1"
	extensionslisters "k8s.io/client-go/listers/extensions/v1beta1"
	"k8s.io/client-go/pkg/api/v1"
	extensions "k8s.io/client-go/pkg/apis/extensions/v1beta1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/kubernetes/pkg/controller"
)

const (
	configMapName    = "%s-proxy-nginx-config"
	tcpConfigMapName = "%s-proxy-nginx-tcp"
	udpConfigMapName = "%s-proxy-nginx-udp"
	proxyNameSuffix  = "-proxy-nginx"
	// ingress controller use this port to export metrics and pprof information
	ingressControllerPort = 8282
	proxyName             = "nginx"
)

var (
	// controllerKind contains the schema.GroupVersionKind for this controller type.
	controllerKind = netv1alpha1.SchemeGroupVersion.WithKind(netv1alpha1.LoadBalancerKind)
)

func init() {
	proxy.RegisterPlugin(proxyName, NewNginx())
}

var _ proxy.Plugin = &nginx{}

type nginx struct {
	initialized           bool
	image                 string
	defaultHTTPbackend    string
	defaultSSLCertificate string

	client    kubernetes.Interface
	tprclient tprclient.Interface

	helper *controllerutil.Helper

	lbLister        netlisters.LoadBalancerLister
	dLister         extensionslisters.DeploymentLister
	podLister       corelisters.PodLister
	lbListerSynced  cache.InformerSynced
	dListerSynced   cache.InformerSynced
	podListerSynced cache.InformerSynced

	queue workqueue.RateLimitingInterface
}

// NewNginx creates a new nginx proxy plugin
func NewNginx() proxy.Plugin {
	return &nginx{}
}

func (f *nginx) Init(cfg config.Configuration, sif informers.SharedInformerFactory) {
	if f.initialized {
		return
	}
	f.initialized = true

	log.Info("Initialize the nginx proxy")
	// set config
	f.defaultHTTPbackend = cfg.Proxies.DefaultHTTPBackend
	f.defaultSSLCertificate = cfg.Proxies.DefaultSSLCertificate
	f.image = cfg.Proxies.Nginx.Image
	f.client = cfg.Client
	f.tprclient = cfg.TPRClient

	// initialize controller
	lbInformer := sif.Networking().V1alpha1().LoadBalancer()
	dInformer := sif.Extensions().V1beta1().Deployments()
	podInfomer := sif.Core().V1().Pods()

	f.lbLister = lbInformer.Lister()
	f.dLister = dInformer.Lister()
	f.podLister = podInfomer.Lister()

	f.queue = workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "proxy-nginx")
	f.helper = controllerutil.NewHelperForKeyFunc(&netv1alpha1.LoadBalancer{}, f.queue, f.syncLoadBalancer, controllerutil.PassthroughKeyFunc)

	dInformer.Informer().AddEventHandler(lbutil.NewEventHandlerForDeployment(f.lbLister, f.dLister, f.helper, f.deploymentFiltered))
	podInfomer.Informer().AddEventHandler(lbutil.NewEventHandlerForSyncStatusWithPod(f.lbLister, f.podLister, f.helper, f.podFiltered))
}

func (f *nginx) Run(stopCh <-chan struct{}) {
	workers := 1
	if !f.initialized {
		log.Panic("Please initialize proxy before you run it")
		return
	}

	defer utilruntime.HandleCrash()

	log.Info("Starting nginx proxy", log.Fields{"workers": workers, "image": f.image, "default-http-backend": f.defaultHTTPbackend})
	defer log.Info("Shutting down nginx proxy")

	if err := f.ensureDefaultHTTPBackend(); err != nil {
		log.Error("ensure default http backend service error", log.Fields{"err": err})
		return
	}

	// lb controller has waited all the informer synced
	// there is no need to wait again here

	defer func() {
		log.Info("Shutting down nginx proxy")
		f.helper.ShutDown()
	}()

	f.helper.Run(workers, stopCh)

	<-stopCh

}

func (f *nginx) selector(lb *netv1alpha1.LoadBalancer) labels.Set {
	return labels.Set{
		netv1alpha1.LabelKeyCreatedBy: fmt.Sprintf(netv1alpha1.LabelValueFormatCreateby, lb.Namespace, lb.Name),
		netv1alpha1.LabelKeyProxy:     proxyName,
	}
}

// filter Deployment that controller does not care
func (f *nginx) deploymentFiltered(obj *extensions.Deployment) bool {
	return f.filteredByLabel(obj)
}

func (f *nginx) podFiltered(obj *v1.Pod) bool {
	return f.filteredByLabel(obj)
}

func (f *nginx) filteredByLabel(obj metav1.ObjectMetaAccessor) bool {
	// obj.Labels
	selector := labels.Set{netv1alpha1.LabelKeyProxy: proxyName}.AsSelector()
	match := selector.Matches(labels.Set(obj.GetObjectMeta().GetLabels()))

	return !match
}

func (f *nginx) OnSync(lb *netv1alpha1.LoadBalancer) {
	if lb.Spec.Proxy.Type != netv1alpha1.ProxyTypeNginx {
		// It is not my responsible
		return
	}
	log.Info("Syncing proxy, triggered by lb controller", log.Fields{"lb": lb.Name, "namespace": lb.Namespace})
	f.helper.Enqueue(lb)
}

// TODO use event
// sync deployment with loadbalancer
// the obj will be *netv1alpha1.LoadBalancer
func (f *nginx) syncLoadBalancer(obj interface{}) error {
	lb, ok := obj.(*netv1alpha1.LoadBalancer)
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
		log.Debug("Finished syncing nginx proxy", log.Fields{"lb": key, "usedTime": time.Since(startTime)})
	}()

	nlb, err := f.lbLister.LoadBalancers(lb.Namespace).Get(lb.Name)
	if errors.IsNotFound(err) {
		log.Warn("LoadBalancer has been deleted, clean up proxy", log.Fields{"lb": key})

		return f.cleanup(lb)
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

	lb, err = f.clone(nlb)
	if err != nil {
		return err
	}

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

func (f *nginx) getDeploymentsForLoadBalancer(lb *netv1alpha1.LoadBalancer) ([]*extensions.Deployment, error) {

	// construct selector
	selector := f.selector(lb).AsSelector()

	// list all
	dList, err := f.dLister.Deployments(lb.Namespace).List(selector)
	if err != nil {
		return nil, err
	}

	// If any adoptions are attempted, we should first recheck for deletion with
	// an uncached quorum read sometime after listing deployment (see kubernetes#42639).
	canAdoptFunc := controller.RecheckDeletionTimestamp(func() (metav1.Object, error) {
		// fresh lb
		fresh, err := f.tprclient.NetworkingV1alpha1().LoadBalancers(lb.Namespace).Get(lb.Name, metav1.GetOptions{})
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
func (f *nginx) sync(lb *netv1alpha1.LoadBalancer, dps []*extensions.Deployment) error {
	desiredDeploy := f.GenerateDeployment(lb)

	// update
	var err error
	updated := false
	activeDeploy := desiredDeploy

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
			log.Info("Scale unexpected proxy replicas to zero", log.Fields{"d.name": dp.Name, "lb.name": lb.Name})
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
			log.Info("Sync nginx for lb", log.Fields{"d.name": dp.Name, "lb.name": lb.Name})
			_, err = f.client.ExtensionsV1beta1().Deployments(lb.Namespace).Update(copyDp)
			if err != nil {
				return err
			}
		}
		activeDeploy = copyDp
	}

	// len(dps) == 0 or no deployment's name match desired deployment
	if !updated {
		// create deployment
		log.Info("Create nginx for lb", log.Fields{"d.name": desiredDeploy.Name, "lb.name": lb.Name})
		_, err = f.client.ExtensionsV1beta1().Deployments(lb.Namespace).Create(desiredDeploy)
		if err != nil {
			return err
		}
	}

	err = f.ensureConfigMaps(lb)
	if err != nil {
		return err
	}

	// update status
	return f.syncStatus(lb, activeDeploy)
}

func (f *nginx) ensureDeployment(desiredDeploy, oldDeploy *extensions.Deployment) (*extensions.Deployment, bool, error) {
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
	// ensure image
	copyDp.Spec.Template.Spec.Containers[0].Image = desiredDeploy.Spec.Template.Spec.Containers[0].Image
	// ensure nodeaffinity
	copyDp.Spec.Template.Spec.Affinity.NodeAffinity = desiredDeploy.Spec.Template.Spec.Affinity.NodeAffinity

	// check if changed
	nodeAffinityChanged := !reflect.DeepEqual(copyDp.Spec.Template.Spec.Affinity.NodeAffinity, oldDeploy.Spec.Template.Spec.Affinity.NodeAffinity)
	imageChanged := copyDp.Spec.Template.Spec.Containers[0].Image != oldDeploy.Spec.Template.Spec.Containers[0].Image
	labelChanged := !reflect.DeepEqual(copyDp.Labels, oldDeploy.Labels)
	replicasChanged := *(copyDp.Spec.Replicas) != *(oldDeploy.Spec.Replicas)

	changed := labelChanged || replicasChanged || nodeAffinityChanged || imageChanged
	if changed {
		log.Info("Abount to correct nginx proxy", log.Fields{
			"dp.name":             copyDp.Name,
			"labelChanged":        labelChanged,
			"replicasChanged":     replicasChanged,
			"nodeAffinityChanged": nodeAffinityChanged,
			"imageChanged":        imageChanged,
		})
	}

	return copyDp, changed, nil
}

// cleanup deployment and other resource controlled by lb proxy
func (f *nginx) cleanup(lb *netv1alpha1.LoadBalancer) error {

	selector := f.selector(lb)

	ds, err := f.getDeploymentsForLoadBalancer(lb)
	if err != nil {
		return err
	}

	policy := metav1.DeletePropagationForeground
	gracePeriodSeconds := int64(30)

	for _, d := range ds {
		err = f.client.ExtensionsV1beta1().Deployments(d.Namespace).Delete(d.Name, &metav1.DeleteOptions{
			GracePeriodSeconds: &gracePeriodSeconds,
			PropagationPolicy:  &policy,
		})
		if err != nil {
			log.Warn("Cleanup proxy error", log.Fields{"ns": d.Namespace, "d.name": d.Name, "err": err})
			return err
		}
	}

	// clean up config map
	err = f.client.CoreV1().ConfigMaps(lb.Namespace).DeleteCollection(nil, metav1.ListOptions{
		LabelSelector: selector.String(),
	})
	if err != nil {
		log.Warn("Cleanup ConfigMap error", log.Fields{"err": err})
		return err
	}

	// clean up ingress
	selector = labels.Set{
		// createdby ingressClass
		netv1alpha1.LabelKeyCreatedBy: fmt.Sprintf(netv1alpha1.LabelValueFormatCreateby, lb.Namespace, lb.Name),
	}
	ingresses, err := f.client.ExtensionsV1beta1().Ingresses(metav1.NamespaceAll).List(metav1.ListOptions{
		LabelSelector: selector.String(),
	})

	if err != nil {
		log.Warn("Cleanup Ingress error", log.Fields{"err": err})
		return err
	}

	for _, ingress := range ingresses.Items {
		err = f.client.ExtensionsV1beta1().Ingresses(ingress.Namespace).Delete(ingress.Name, &metav1.DeleteOptions{})
		if err != nil {
			log.Warn("Cleanup Ingress error", log.Fields{"err": err})
			return err
		}
	}

	return nil
}

func (f *nginx) GenerateDeployment(lb *netv1alpha1.LoadBalancer) *extensions.Deployment {
	terminationGracePeriodSeconds := int64(30)
	hostNetwork := false
	replicas, needNodeAffinity := lbutil.CalculateReplicas(lb)

	if lb.Spec.Type == netv1alpha1.LoadBalancerTypeExternal {
		hostNetwork = true
	}

	labels := f.selector(lb)

	// run on this node
	nodeAffinity := &v1.NodeAffinity{
		RequiredDuringSchedulingIgnoredDuringExecution: &v1.NodeSelector{
			NodeSelectorTerms: []v1.NodeSelectorTerm{
				{
					MatchExpressions: []v1.NodeSelectorRequirement{
						{
							Key:      fmt.Sprintf(netv1alpha1.UniqueLabelKeyFormat, lb.Namespace, lb.Name),
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
					MatchLabels: map[string]string{
						netv1alpha1.LabelKeyProxy: proxyName,
					},
				},
				TopologyKey: metav1.LabelHostname,
			},
		},
	}

	t := true

	deploy := &extensions.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:   lb.Name + proxyNameSuffix + "-" + lbutil.RandStringBytesRmndr(5),
			Labels: labels,
			Annotations: map[string]string{
				"prometheus.io/port":   strconv.Itoa(ingressControllerPort),
				"prometheus.io/scrape": "true",
			},
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
						// don't co-locate pods of this deployment in same node
						PodAntiAffinity: podAffinity,
					},
					Tolerations: toleration.GenerateTolerations(),
					Containers: []v1.Container{
						{
							Name:            "ingress-nginx-controller",
							Image:           f.image,
							ImagePullPolicy: v1.PullAlways,
							Resources:       lb.Spec.Proxy.Resources,
							Ports: []v1.ContainerPort{
								{
									ContainerPort: 80,
								},
								{
									ContainerPort: 443,
								},
								{
									ContainerPort: ingressControllerPort,
								},
							},
							Env: []v1.EnvVar{
								{
									Name: "POD_NAME",
									ValueFrom: &v1.EnvVarSource{
										FieldRef: &v1.ObjectFieldSelector{
											FieldPath: "metadata.name",
										},
									},
								},
								{
									Name: "POD_NAMESPACE",
									ValueFrom: &v1.EnvVarSource{
										FieldRef: &v1.ObjectFieldSelector{
											FieldPath: "metadata.namespace",
										},
									},
								},
							},
							// TODO
							Args: []string{
								"/nginx-ingress-controller",
								"--default-backend-service=" + fmt.Sprintf("%s/%s", defaultHTTPBackendNamespace, defaultHTTPBackendName),
								"--ingress-class=" + fmt.Sprintf(netv1alpha1.LabelValueFormatCreateby, lb.Namespace, lb.Name),
								"--configmap=" + fmt.Sprintf("%s/"+configMapName, lb.Namespace, lb.Name),
								"--tcp-services-configmap=" + fmt.Sprintf("%s/"+tcpConfigMapName, lb.Namespace, lb.Name),
								"--udp-services-configmap=" + fmt.Sprintf("%s/"+udpConfigMapName, lb.Namespace, lb.Name),
								"--healthz-port=" + strconv.Itoa(ingressControllerPort),
								"--health-check-path=" + "/healthz",
							},
							ReadinessProbe: &v1.Probe{
								Handler: v1.Handler{
									HTTPGet: &v1.HTTPGetAction{
										Path:   "/healthz",
										Port:   intstr.FromInt(80),
										Scheme: v1.URISchemeHTTP,
									},
								},
							},
							LivenessProbe: &v1.Probe{
								Handler: v1.Handler{
									HTTPGet: &v1.HTTPGetAction{
										Path:   "/healthz",
										Port:   intstr.FromInt(80),
										Scheme: v1.URISchemeHTTP,
									},
								},
							},
						},
					},
				},
			},
		},
	}

	if needNodeAffinity {
		// decide running on which node
		deploy.Spec.Template.Spec.Affinity.NodeAffinity = nodeAffinity
	}

	if f.defaultSSLCertificate != "" {
		deploy.Spec.Template.Spec.Containers[0].Args = append(
			deploy.Spec.Template.Spec.Containers[0].Args,
			"--default-ssl-certificate="+f.defaultSSLCertificate,
		)
	}

	return deploy
}

func (f *nginx) clone(lb *netv1alpha1.LoadBalancer) (*netv1alpha1.LoadBalancer, error) {
	lbi, err := scheme.Scheme.DeepCopy(lb)
	if err != nil {
		log.Error("Unable to deepcopy loadbalancer", log.Fields{"lb.name": lb.Name, "err": err})
		return nil, err
	}

	nlb, ok := lbi.(*netv1alpha1.LoadBalancer)
	if !ok {
		nerr := fmt.Errorf("expected loadbalancer, got %#v", lbi)
		log.Error(nerr)
		return nil, err
	}
	return nlb, nil
}
