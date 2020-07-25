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

package config

import (
	"strings"

	"github.com/caicloud/clientset/kubernetes"
	"github.com/caicloud/loadbalancer-controller/pkg/toleration"
	"github.com/spf13/pflag"
)

const (
	defaultIpvsdrImage             = "cargo.caicloud.io/caicloud/loadbalancer-provider-ipvsdr:v0.3.2"
	defaultAzureProviderImage      = "cargo.caicloud.io/caicloud/loadbalancer-provider-azure:v0.3.2"
	defaultHTTPBackendImage        = "cargo.caicloud.io/caicloud/default-http-backend:v0.1.0"
	defaultNginxIngressImage       = "cargo.caicloud.io/caicloud/nginx-ingress-controller:0.12.0"
	defaultIngressSidecarImage     = "cargo.caicloud.io/caicloud/loadbalancer-provider-ingress:v0.3.2"
	defaultIngressAnnotationPrefix = "ingress.kubernetes.io"
)

type additionalTolerations []string

func (a *additionalTolerations) Set(value string) error {
	values := strings.Split(value, ",")
	if len(values) == 0 {
		return nil
	}
	*a = append(*a, values...)
	// add additional keys
	toleration.AddAdditionalTolerationKeys(*a)
	return nil
}

func (a *additionalTolerations) String() string {
	return strings.Join(*a, ",")
}

func (a *additionalTolerations) Type() string {
	return "AdditionalTolerations"
}

// Configuration contains the global config of controller
type Configuration struct {
	Client                kubernetes.Interface
	AdditionalTolerations additionalTolerations
	Proxies               Proxies
	Providers             Providers
}

// Proxies contains all cli flags of proxies
type Proxies struct {
	Sidecar string
	Nginx   ProxyNginx
}

// ProxyNginx contains all cli flags of nginx proxy
type ProxyNginx struct {
	Image                 string
	DefaultHTTPBackend    string
	AnnotationPrefix      string
	DefaultSSLCertificate string
}

// Providers contains all cli flags of providers
type Providers struct {
	Ipvsdr ProviderIpvsdr
	Azure  ProviderAzure
	F5     ProviderF5
}

// ProviderIpvsdr contains all cli flags of ipvsdr providers
type ProviderIpvsdr struct {
	Image            string
	NodeIPLabel      string
	NodeIPAnnotation string
}

// ProviderAzure contains all cli flags of azure providers
type ProviderAzure struct {
	Image string
}

// ProviderF5 contains all cli flags of f5 providers
type ProviderF5 struct {
	Image string
}

// AddFlags add flags to app
func (c *Configuration) AddFlags(fs *pflag.FlagSet) {

	fs.Var(&c.AdditionalTolerations, "additional-tolerations", "A comma separated list of k8s `TolerationKeys`")

	fs.StringVar(&c.Proxies.Sidecar, "proxy-sidecar", defaultIngressSidecarImage, "`Image` of ingress controller sidecar")

	fs.StringVar(&c.Proxies.Nginx.Image, "proxy-nginx", defaultNginxIngressImage, "`Image` of nginx ingress controller image")
	fs.StringVar(&c.Proxies.Nginx.DefaultHTTPBackend, "default-http-backend", defaultHTTPBackendImage, "Default http backend `Image` for ingress controller")
	fs.StringVar(&c.Proxies.Nginx.DefaultSSLCertificate, "default-ssl-certificate", "", "Name of the secret that contains a SSL `certificate` to be used as default for a HTTPS catch-all server")
	fs.StringVar(&c.Proxies.Nginx.AnnotationPrefix, "proxy-nginx-annotation-prefix", defaultIngressAnnotationPrefix, "Prefix of ingress annotation")

	fs.StringVar(&c.Providers.Ipvsdr.Image, "provider-ipvsdr", defaultIpvsdrImage, "`Image` of ipvsdr provider")
	fs.StringVar(&c.Providers.F5.Image, "provider-f5", defaultIpvsdrImage, "`Image` of ipvsdr provider")
	fs.StringVar(&c.Providers.Ipvsdr.NodeIPLabel, "nodeip-label", "", "tell provider which label of node stores node ip")
	fs.StringVar(&c.Providers.Ipvsdr.NodeIPAnnotation, "nodeip-annotation", "", "tell provider which annotation of node stores node ip")

	fs.StringVar(&c.Providers.Azure.Image, "provider-azure", defaultAzureProviderImage, "`Image` of azure provider")

}
