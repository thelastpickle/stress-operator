package stresscontext

import (
	"context"
	"github.com/thelastpickle/stress-operator/pkg/apis/thelastpickle/v1alpha1"
	"github.com/thelastpickle/stress-operator/pkg/casskop"
	"github.com/thelastpickle/stress-operator/pkg/monitoring"
	"github.com/thelastpickle/stress-operator/test"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"testing"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	prometheusv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	corev1 "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1"
	i8ly "github.com/integr8ly/grafana-operator/v3/pkg/apis/integreatly/v1alpha1"
)

type fakeDiscoveryClient struct {}

func (fdc *fakeDiscoveryClient) KindExists(apiVersion string, kind string) (bool, error) {
	return true, nil
}

func setupReconcile(t *testing.T, state ...runtime.Object) (*ReconcileStressContext, reconcile.Result) {
	cl := fake.NewFakeClient(state...)
	r := &ReconcileStressContext{client: cl, scheme: testScheme}
	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      contextName,
			Namespace: namespace,
		},
	}
	res, err := r.Reconcile(req)
	if err != nil {
		t.Fatalf("reconcile: (%v)", err)
	}

	return r, res
}

func setupReconcileWithRequeue(t *testing.T, state ...runtime.Object) *ReconcileStressContext {
	r, res := setupReconcile(t, state...)

	// Check the result of reconciliation to make sure it has the desired state.
	if !res.Requeue {
		t.Error("reconcile did not requeue request as expected")
	}

	return r
}

func setupReconcileWithoutRequeue(t *testing.T, state ...runtime.Object) *ReconcileStressContext {
	r, res := setupReconcile(t, state...)

	if res.Requeue {
		t.Error("did not expect reconcile to requeue the request")
	}

	return r
}

var (
	name          = "stress-controller"
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
	casskop.Init(fdc)

	test.InitScheme(t)

	t.Run("CreatePrometheus", testCreatePrometheus)
	t.Run("CreatePrometheusService", testCreatePrometheusService)
	t.Run("CreateServiceMonitor", testCreateServiceMonitor)
	t.Run("CreateGrafana", testCreateGrafana)
	t.Run("CreateGrafanaDataSource", testCreateGrafanaDataSource)
}

func testCreatePrometheus(t *testing.T) {
	ctx := createStressContext()

	objs := []runtime.Object{ctx}

	r := setupReconcileWithRequeue(t, objs...)

	if _, err := monitoring.GetPrometheus(namespace, r.client); err != nil {
		t.Errorf("get prometheus: (%v)", err)
	}

	sa := &corev1.ServiceAccount{}
	if err := r.client.Get(context.TODO(), types.NamespacedName{Namespace: namespace, Name: monitoring.PrometheusName}, sa); err != nil {
		t.Errorf("get prometheus service account: (%v)", err)
	}

	role := &rbac.Role{}
	if err := r.client.Get(context.TODO(), types.NamespacedName{Namespace: namespace, Name: monitoring.PrometheusName}, role); err != nil {
		t.Errorf("get prometheus role: (%v)", err)
	}

	roleBinding := &rbac.RoleBinding{}
	if err := r.client.Get(context.TODO(), types.NamespacedName{Namespace: namespace, Name: monitoring.PrometheusName}, roleBinding); err != nil {
		t.Errorf("get prometheus role binding: (%v)", err)
	}
}

func testCreatePrometheusService(t *testing.T) {
	ctx := createStressContext()
	prometheus := createPrometheus()

	objs := []runtime.Object{ctx, prometheus}

	r := setupReconcileWithRequeue(t, objs...)

	svc := &corev1.Service{}
	if err := r.client.Get(context.TODO(), types.NamespacedName{Namespace: namespace, Name: monitoring.PrometheusName}, svc); err != nil {
		t.Errorf("get prometheus service: (%v)", err)
	}
}

func testCreateServiceMonitor(t *testing.T) {
	ctx := createStressContext()
	prometheus := createPrometheus()
	prometheusService := createPrometheusService()

	objs := []runtime.Object{ctx, prometheus, prometheusService}

	r := setupReconcileWithRequeue(t, objs...)

	if _, err := monitoring.GetServiceMonitor(namespace, r.client); err != nil {
		t.Errorf("get service monitor: (%v)", err)
	}
}

func testCreateGrafana(t *testing.T) {
	ctx := createStressContext()
	prometheus := createPrometheus()
	prometheusService := createPrometheusService()
	serviceMonitor :=  createServiceMonitor()

	objs := []runtime.Object{ctx, prometheus, prometheusService, serviceMonitor}

	r := setupReconcileWithRequeue(t, objs...)

	if _, err := monitoring.GetGrafana(namespace, r.client); err != nil {
		t.Errorf("get grafana: (%v)", err)
	}
}

func testCreateGrafanaDataSource(t *testing.T) {
	ctx := createStressContext()
	prometheus := createPrometheus()
	prometheusService := createPrometheusService()
	serviceMonitor :=  createServiceMonitor()
	grafana := createGrafana()

	objs := []runtime.Object{ctx, prometheus, prometheusService, serviceMonitor, grafana}

	r := setupReconcileWithRequeue(t, objs...)

	if _, err := monitoring.GetDataSource(namespace, r.client); err != nil {
		t.Errorf("get grafana data source: (%v)", err)
	}
}

func reconcileNonDefaultContext(t *testing.T) {

}

func createStressContext() *v1alpha1.StressContext {
	return &v1alpha1.StressContext{
		ObjectMeta: metav1.ObjectMeta{
			Name:      contextName,
			Namespace: namespace,
		},
		Spec: v1alpha1.StressContextSpec{
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

func createServiceMonitor() *prometheusv1.ServiceMonitor {
	return &prometheusv1.ServiceMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name: monitoring.ServiceMonitorName,
		},
	}
}

func createGrafana() *i8ly.Grafana {
	return &i8ly.Grafana{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name: monitoring.GrafanaName,
		},
	}
}