package tlpstresscontext

import (
	"context"
	v1alpha1 "github.com/jsanda/tlp-stress-operator/pkg/apis/thelastpickle/v1alpha1"
	"github.com/jsanda/tlp-stress-operator/pkg/monitoring"
	"github.com/jsanda/tlp-stress-operator/test"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"testing"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	prometheusv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	corev1 "k8s.io/api/core/v1"
)

type fakeDiscoveryClient struct {}

func (fdc *fakeDiscoveryClient) KindExists(apiVersion string, kind string) (bool, error) {
	return true, nil
}

func setupReconcile(t *testing.T, state ...runtime.Object) (*ReconcileTLPStressContext, reconcile.Result) {
	cl := fake.NewFakeClient(state...)
	r := &ReconcileTLPStressContext{client: cl, scheme: testScheme}
	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      name,
			Namespace: namespace,
		},
	}
	res, err := r.Reconcile(req)
	if err != nil {
		t.Fatalf("reconcile: (%v)", err)
	}

	return r, res
}

func setupReconcileWithRequeue(t *testing.T, state ...runtime.Object) *ReconcileTLPStressContext {
	r, res := setupReconcile(t, state...)

	// Check the result of reconciliation to make sure it has the desired state.
	if !res.Requeue {
		t.Error("reconcile did not requeue request as expected")
	}

	return r
}

func setupReconcileWithoutRequeue(t *testing.T, state ...runtime.Object) *ReconcileTLPStressContext {
	r, res := setupReconcile(t, state...)

	if res.Requeue {
		t.Error("did not expect reconcile to requeue the request")
	}

	return r
}

var (
	name          = "tlpstress-controller"
	namespace     = "tlpstress"
	namespaceName = types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}
	fdc          = &fakeDiscoveryClient{}
	testScheme   = scheme.Scheme
)

func TestReconcile(t *testing.T) {
	monitoring.Init(fdc)
	monitoring.Init(fdc)

	test.InitScheme(t)

	t.Run("CreatePrometheus", testCreatePrometheus)
	t.Run("CreatePrometheusService", testCreatePrometheusService)
	t.Run("CreateServiceMonitor", testCreateServiceMonitor)
	t.Run("CreateGrafana", testCreateGrafana)
}

func testCreatePrometheus(t *testing.T) {
	ctx := createTLPStressContext()

	objs := []runtime.Object{ctx}

	r := setupReconcileWithRequeue(t, objs...)

	if _, err := monitoring.GetPrometheus(namespace, r.client); err != nil {
		t.Errorf("get prometheus: (%v)", err)
	}
}

func testCreatePrometheusService(t *testing.T) {
	ctx := createTLPStressContext()
	prometheus := createPrometheus()

	objs := []runtime.Object{ctx, prometheus}

	r := setupReconcileWithRequeue(t, objs...)

	svc := &corev1.Service{}
	if err := r.client.Get(context.TODO(), types.NamespacedName{Namespace: namespace, Name: monitoring.PrometheusName}, svc); err != nil {
		t.Errorf("get prometheus service: (%v)", err)
	}
}

func testCreateServiceMonitor(t *testing.T) {
	ctx := createTLPStressContext()
	prometheus := createPrometheus()
	prometheusService := createPrometheusService()

	objs := []runtime.Object{ctx, prometheus, prometheusService}

	r := setupReconcileWithRequeue(t, objs...)

	if _, err := monitoring.GetServiceMonitor(namespace, r.client); err != nil {
		t.Errorf("get service monitor: (%v)", err)
	}
}

func testCreateGrafana(t *testing.T) {
	
}

func createTLPStressContext() *v1alpha1.TLPStressContext {
	return &v1alpha1.TLPStressContext{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: v1alpha1.TLPStressContextSpec{
			InstallPrometheus: true,
			InstallGrafana: true,
		},
	}
}

func createPrometheus() *prometheusv1.Prometheus {
	return &prometheusv1.Prometheus{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name: monitoring.PrometheusName,
		},
	}
}

func createPrometheusService() *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      monitoring.PrometheusName,
		},
	}
}