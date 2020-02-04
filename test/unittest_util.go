package test

import (
	"github.com/jsanda/stress-operator/pkg/apis"
	"github.com/jsanda/stress-operator/pkg/casskop"
	"github.com/jsanda/stress-operator/pkg/monitoring"
	"k8s.io/client-go/kubernetes/scheme"
	"testing"
)

type fakeDiscoveryClient struct {}

func NewFakeDiscoveryClient() *fakeDiscoveryClient {
	return &fakeDiscoveryClient{}
}

func (fdc *fakeDiscoveryClient) KindExists(apiVersion string, kind string) (bool, error) {
	return true, nil
}

func InitScheme(t *testing.T) {
	if err := apis.AddToScheme(scheme.Scheme); err != nil {
		t.FailNow()
	}
	if err := monitoring.AddToScheme(scheme.Scheme); err != nil {
		t.FailNow()
	}
	if err := casskop.AddToScheme(scheme.Scheme); err != nil {
		t.FailNow()
	}
}