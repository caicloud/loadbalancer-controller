package kong

import (
	// "github.com/caicloud/clientset/kubernetes"
	// apiextv1beta1 "github.com/caicloud/clientset/pkg/apis/apiextensions/v1beta1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apiextv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	log "k8s.io/klog"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	kongCrdGroup = "configuration.konghq.com"
)

var kongCrds = []apiextv1beta1.CustomResourceDefinition{
	kongclusterplugins,
	kongconsumers,
	kongcredentials,
	kongingresses,
	kongplugins,
	tcpingresses,
}

func installKongCrds() error {
	// create client 
	// caicloud clientset can't create crd with AdditionalPrinterColumns and Validation
	kubeconfig, err := clientcmd.BuildConfigFromFlags("","")
	if err != nil {
		log.Errorf("build kubeconfig error for kong crd, error %v", err)
		return err
	}
	client, err := apiextensionsv1beta1.NewForConfig(kubeconfig)
	if err != nil {
		log.Errorf("create apiextensionsv1beta1 clientset error %v", err)
		return err
	}
	// all crds
	for _, crd := range kongCrds {
		//
		_, err := client.ApiextensionsV1beta1().CustomResourceDefinitions().Get(crd.Name, metav1.GetOptions{})
		if err != nil {
			if k8serrors.IsNotFound(err) {
				// create the crd
				log.Infof("Create kong crd %v", crd.Name)
				if _, err := client.ApiextensionsV1beta1().CustomResourceDefinitions().Create(&crd); err != nil {
					log.Errorf("Create crd %v error %v", crd.Name, err)
					return err
				}
			} else {
				log.Errorf("Get crd %v error %v", crd.Name, err)
				return err
			}
		}
		log.Infof("Get crd %v installed", crd.Name)
	}
	return nil
}

// kongclusterplugins
var kongclusterplugins = apiextv1beta1.CustomResourceDefinition{
	ObjectMeta: metav1.ObjectMeta{
		Name: "kongclusterplugins.configuration.konghq.com",
	},
	Spec: apiextv1beta1.CustomResourceDefinitionSpec{
		AdditionalPrinterColumns: []apiextv1beta1.CustomResourceColumnDefinition{
			{
				JSONPath: ".plugin",
				Description: "Name of the plugin",
				Name: "Plugin-Type",
				Type: "string",
			},
			{
				JSONPath: ".metadata.creationTimestamp",
				Description: "Age",
				Name: "Age",
				Type: "date",
			},
			{
				JSONPath: ".disabled",
				Description: "Indicates if the plugin is disabled",
				Name: "Disabled",
				Priority: 1,
				Type: "boolean",
			},
			{
				JSONPath: ".config",
				Description: "Configuration of the plugin",
				Name: "Config",
				Priority: 1,
				Type: "string",
			},
		},
		Group: kongCrdGroup,
		Names: apiextv1beta1.CustomResourceDefinitionNames{
			Kind: "KongClusterPlugin",
			Plural: "kongclusterplugins",
			ShortNames: []string{
				"kcp",
			},
		},
		Scope: apiextv1beta1.ClusterScoped,
		// Validation
		Version: "v1",
	},
}

// kongconsumers
var kongconsumers = apiextv1beta1.CustomResourceDefinition{
	ObjectMeta: metav1.ObjectMeta{
		Name: "kongconsumers.configuration.konghq.com",
	},
	Spec: apiextv1beta1.CustomResourceDefinitionSpec{
		AdditionalPrinterColumns: []apiextv1beta1.CustomResourceColumnDefinition{
			{
				JSONPath: ".username",
				Description: "Username of a Kong Consumer",
				Name: "Username",
				Type: "string",
			},
			{
				JSONPath: ".metadata.creationTimestamp",
				Description: "Age",
				Name: "Age",
				Type: "date",
			},
		},
		Group: kongCrdGroup,
		Names: apiextv1beta1.CustomResourceDefinitionNames{
			Kind: "KongConsumer",
			Plural: "kongconsumers",
			ShortNames: []string{
				"kc",
			},
		},
		Scope: apiextv1beta1.NamespaceScoped,
		// Validation
		Version: "v1",
	},
}

// kongcredentials
var kongcredentials = apiextv1beta1.CustomResourceDefinition{
	ObjectMeta: metav1.ObjectMeta{
		Name: "kongcredentials.configuration.konghq.com",
	},
	Spec: apiextv1beta1.CustomResourceDefinitionSpec{
		AdditionalPrinterColumns: []apiextv1beta1.CustomResourceColumnDefinition{
			{
				JSONPath: ".type",
				Description: "Type of credential",
				Name: "Credential-type",
				Type: "string",
			},
			{
				JSONPath: ".metadata.creationTimestamp",
				Description: "Age",
				Name: "Age",
				Type: "date",
			},
			{
				JSONPath: ".consumerRef",
				Description: "Owner of the credential",
				Name: "Consumer-Ref",
				Type: "string",
			},
		},
		Group: kongCrdGroup,
		Names: apiextv1beta1.CustomResourceDefinitionNames{
			Kind: "KongCredential",
			Plural: "kongcredentials",
		},
		Scope: apiextv1beta1.NamespaceScoped,
		// Validation
		Version: "v1",
	},
}

// kongingresses
var kongingresses = apiextv1beta1.CustomResourceDefinition{
	ObjectMeta: metav1.ObjectMeta{
		Name: "kongingresses.configuration.konghq.com",
	},
	Spec: apiextv1beta1.CustomResourceDefinitionSpec{
		Group: kongCrdGroup,
		Names: apiextv1beta1.CustomResourceDefinitionNames{
			Kind: "KongIngress",
			Plural: "kongingresses",
			ShortNames: []string{
				"ki",
			},
		},
		Scope: apiextv1beta1.NamespaceScoped,
		// Validation
		Version: "v1",
	},
}

// kongplugins
var kongplugins = apiextv1beta1.CustomResourceDefinition{
	ObjectMeta: metav1.ObjectMeta{
		Name: "kongplugins.configuration.konghq.com",
	},
	Spec: apiextv1beta1.CustomResourceDefinitionSpec{
		AdditionalPrinterColumns: []apiextv1beta1.CustomResourceColumnDefinition{
			{
				JSONPath: ".plugin",
				Description: "Name of the plugin",
				Name: "Plugin-Type",
				Type: "string",
			},
			{
				JSONPath: ".metadata.creationTimestamp",
				Description: "Age",
				Name: "Age",
				Type: "date",
			},
			{
				JSONPath: ".disabled",
				Description: "Indicates if the plugin is disabled",
				Name: "Disabled",
				Priority: 1,
				Type: "boolean",
			},
			{
				JSONPath: ".config",
				Description: "Configuration of the plugin",
				Name: "Config",
				Priority: 1,
				Type: "string",
			},
		},
		Group: kongCrdGroup,
		Names: apiextv1beta1.CustomResourceDefinitionNames{
			Kind: "KongPlugin",
			Plural: "kongplugins",
			ShortNames: []string{
				"kp",
			},
		},
		Scope: apiextv1beta1.NamespaceScoped,
		// Validation
		Version: "v1",
	},
}

// tcpingresses
var tcpingresses = apiextv1beta1.CustomResourceDefinition{
	ObjectMeta: metav1.ObjectMeta{
		Name: "tcpingresses.configuration.konghq.com",
	},
	Spec: apiextv1beta1.CustomResourceDefinitionSpec{
		AdditionalPrinterColumns: []apiextv1beta1.CustomResourceColumnDefinition{
			{
				JSONPath: ".status.loadBalancer.ingress[*].ip",
				Description: "Address of the load balancer",
				Name: "Address",
				Type: "string",
			},
			{
				JSONPath: ".metadata.creationTimestamp",
				Description: "Age",
				Name: "Age",
				Type: "date",
			},
		},
		Group: kongCrdGroup,
		Names: apiextv1beta1.CustomResourceDefinitionNames{
			Kind: "TCPIngress",
			Plural: "tcpingresses",
		},
		Scope: apiextv1beta1.NamespaceScoped,
		// Validation
		Version: "v1beta1",
	},
}
