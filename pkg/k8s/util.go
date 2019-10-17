package k8s

import (
	"github.com/operator-framework/operator-sdk/pkg/k8sutil"
	"k8s.io/client-go/discovery"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
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