package monitoring

import (
	prometheus "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	i8ly "github.com/integr8ly/grafana-operator/pkg/apis/integreatly/v1alpha1"
	"github.com/jsanda/tlp-stress-operator/pkg/k8s"
	"k8s.io/apimachinery/pkg/runtime"
)

var discoveryClient k8s.DiscoveryClient

func Init(dc k8s.DiscoveryClient) {
	discoveryClient = dc
}

func AddToScheme(scheme *runtime.Scheme) error {
	err := prometheus.AddToScheme(scheme)
	if err != nil {
		return err
	}
	return i8ly.AddToScheme(scheme)
}
