package test

import (
	"github.com/jsanda/tlp-stress-operator/pkg/apis"
	"github.com/jsanda/tlp-stress-operator/pkg/casskop"
	"github.com/jsanda/tlp-stress-operator/pkg/monitoring"
	"k8s.io/client-go/kubernetes/scheme"
	"testing"
)

//type ReconcilerFactory interface {
//	GetReconciler(client client.Client, scheme *runtime.Scheme) reconcile.Reconciler
//}
//
//func Reconcile(
//	t *testing.T,
//	factory ReconcilerFactory,
//	namespacedName types.NamespacedName,
//	state ...runtime.Object) (reconcile.Reconciler, reconcile.Result) {
//
//	cl := fake.NewFakeClient(state...)
//	s := scheme.Scheme
//
//	if err := apis.AddToScheme(s); err != nil {
//		t.FailNow()
//	}
//	if err := monitoring.AddToScheme(s); err != nil {
//		t.FailNow()
//	}
//	if err := casskop.AddToScheme(s); err != nil {
//		t.FailNow()
//	}
//
//	r := factory.GetReconciler(cl, s)
//
//	req := reconcile.Request{NamespacedName: namespacedName}
//
//	res, err := r.Reconcile(req)
//
//	if err != nil {
//		t.Fatalf("reconcile: (%v)", err)
//	}
//
//	return r, res
//}
//
//func ReconcileAndRequeue(
//	t *testing.T,
//	factory ReconcilerFactory,
//	namespacedName types.NamespacedName,
//	state ...runtime.Object) (reconcile.Reconciler, reconcile.Result) {
//
//	r, res := Reconcile(t, factory, namespacedName, state...)
//
//	// Check the result of reconciliation to make sure it has the desired state.
//	if !res.Requeue {
//		t.Error("reconcile did not requeue request as expected")
//	}
//
//	return r, res
//}
//
//func ReconcileAndNoRequeue(
//	t *testing.T,
//	factory ReconcilerFactory,
//	namespacedName types.NamespacedName,
//	state ...runtime.Object) (reconcile.Reconciler, reconcile.Result) {
//
//	r, res := Reconcile(t, factory, namespacedName, state...)
//
//	if res.Requeue {
//		t.Error("did not expect reconcile to requeue the request")
//	}
//
//	return r, res
//}

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