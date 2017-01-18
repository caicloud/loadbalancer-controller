package main

import (
	"fmt"
	"os"
	"reflect"
	"sort"
	"text/template"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/kubernetes/pkg/util/exec"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kubernetes/pkg/util/wait"

	"github.com/golang/glog"
)

type keepalivedController struct {
	clientset  *kubernetes.Clientset
	keepalived *keepalived

	namespace   string
	serviceName string
	podName     string

	template *template.Template
	config   map[string]interface{}
}

func newKeepalivedController(clientset *kubernetes.Clientset, namespace, serviceName, podName string) (*keepalivedController, error) {
	c := &keepalivedController{
		clientset:  clientset,
		keepalived: &keepalived{},

		namespace:   namespace,
		serviceName: serviceName,
		podName:     podName,

		config: make(map[string]interface{}),
	}

	tmpl, err := template.ParseFiles(keepalivedTmpl)
	if err != nil {
		return nil, fmt.Errorf("failed to parse keepalived config template due to: %v", err)
	}
	c.template = tmpl

	return c, nil
}

func (c *keepalivedController) Run(period time.Duration, stopCh <-chan struct{}) {
	go c.keepalived.Start()

	go wait.Until(func() {
		if err := c.Sync(); err != nil {
			glog.Error(err)
		}
	}, period, stopCh)

	<-stopCh
}

func (c *keepalivedController) Stop() error {
	vip := c.config["vip"]
	iface := c.config["iface"]

	defer c.keepalived.Stop()

	return c.removeVIP(iface.(string), vip.(string))
}

func (c *keepalivedController) Sync() error {
	conf, err := c.fetchConfig()
	if err != nil {
		return err
	}

	if reflect.DeepEqual(conf, c.config) {
		return nil
	}
	c.config = conf

	w, err := os.Create(keepalivedCfg)
	if err != nil {
		return fmt.Errorf("failed to create keepalbed config file")
	}
	defer w.Close()

	if err := c.template.Execute(w, conf); err != nil {
		return err
	}

	return c.keepalived.Reload()
}

func (c *keepalivedController) fetchConfig() (conf map[string]interface{}, err error) {
	service, err := c.clientset.Core().Services(c.namespace).Get(c.serviceName, meta_v1.GetOptions{})
	if err != nil {
		return conf, fmt.Errorf("can not get service due to %v", err)
	}
	var vip string
	if service.Annotations != nil {
		vip = service.Annotations[IngressVIPAnnotationKey]
	}
	if vip == "" {
		return conf, fmt.Errorf("no vip has assigned to ingress service")
	}

	endpoint, err := c.clientset.Core().Endpoints(c.namespace).Get(c.serviceName, meta_v1.GetOptions{})
	if err != nil {
		return conf, fmt.Errorf("can not get endpoint due to %v", err)
	}

	peers := []string{}
	for _, subset := range endpoint.Subsets {
		for _, addr := range subset.Addresses {
			peers = append(peers, addr.IP)
		}
		for _, addr := range subset.NotReadyAddresses {
			peers = append(peers, addr.IP)
		}
	}
	sort.Strings(peers)

	pod, err := c.clientset.Core().Pods(c.namespace).Get(c.podName, meta_v1.GetOptions{})
	if err != nil {
		return conf, fmt.Errorf("can not get pod due to %v", err)
	}
	selfIP := pod.Status.PodIP

	neighbors := getNeighbors(selfIP, peers)
	networkInfo, err := getNetworkInfo(selfIP)
	if err != nil {
		return conf, fmt.Errorf("can not get network info due to %v", err)
	}

	conf = make(map[string]interface{})
	conf["iface"] = networkInfo.iface
	conf["selfIP"] = selfIP
	conf["vip"] = vip
	conf["neighbors"] = neighbors
	conf["priority"] = getPriority(selfIP, peers)

	return conf, nil
}

func (c *keepalivedController) removeVIP(iface, vip string) error {
	if iface == "" || vip == "" {
		return nil
	}

	glog.Infof("removing configured VIP %v", vip)
	out, err := exec.New().Command("ip", "addr", "del", vip+"/32", "dev", iface).CombinedOutput()
	if err != nil {
		return fmt.Errorf("error reloading keepalived: %v\n%s", err, out)
	}
	return nil
}
