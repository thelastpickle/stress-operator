package k8s

import (
	"github.com/operator-framework/operator-sdk/pkg/k8sutil"
	"k8s.io/client-go/discovery"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type DiscoveryClient interface {
	KindExists(apiVersion string, kind string) (bool, error)
}

type k8sDiscoveryClient struct {
	discoveryClient *discovery.DiscoveryClient
}

var kdc *k8sDiscoveryClient

func GetDiscoveryClient() (*k8sDiscoveryClient, error) {
	if kdc == nil {
		cfg, err := config.GetConfig()
		if err != nil {
			return nil, err
		}
		// TODO Should we recover from panic and return the error?
		kdc = &k8sDiscoveryClient{ discoveryClient: discovery.NewDiscoveryClientForConfigOrDie(cfg) }
	}
	return kdc, nil
}

func (kdc *k8sDiscoveryClient) KindExists(apiVersion string, kind string) (bool, error) {
	return k8sutil.ResourceExists(kdc.discoveryClient, apiVersion, kind)
}

func CreateServiceAccount(client client.Client, namespace string, name string) error {
	sa := &v1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name: name,
		},
	}
	return CreateResource(client, sa)
}