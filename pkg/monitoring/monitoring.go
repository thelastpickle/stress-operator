package monitoring

import (
	"github.com/jsanda/tlp-stress-operator/pkg/k8s"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var discoveryClient k8s.DiscoveryClient

func Init(dc k8s.DiscoveryClient) {
	discoveryClient = dc
}

func GetKnownTypes() map[schema.GroupVersion][]runtime.Object {
	knownTypes := make(map[schema.GroupVersion][]runtime.Object)

	k, v := getPrometheusTypes()
	knownTypes[k] = v

	k, v = getGrafanaTypes()
	knownTypes[k] = v

	return knownTypes
}
