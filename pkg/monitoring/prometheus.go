package monitoring

import (
	"context"
	"fmt"
	prometheus "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/go-logr/logr"
	api "github.com/thelastpickle/stress-operator/pkg/apis/thelastpickle/v1alpha1"
	"github.com/thelastpickle/stress-operator/pkg/k8s"
	tlp "github.com/thelastpickle/stress-operator/pkg/tlpstress"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"time"
)

const (
	ServiceMonitorKind = "ServiceMonitor"
	PrometheusKind     = "Prometheus"

	PrometheusName     = "stress-prometheus"
	ServiceMonitorName = "stress"

	metricsPort        = "metrics"
)

func GetMetricsService(stress *api.Stress, client client.Client) (*corev1.Service, error) {
	metricsService := &corev1.Service{}
	err := client.Get(context.TODO(), types.NamespacedName{Namespace: stress.Namespace, Name: GetMetricsServiceName(stress)}, metricsService)

	return metricsService, err
}

func CreateMetricsService(stress *api.Stress, client client.Client, log logr.Logger) (reconcile.Result, error) {
	metricsService := newMetricsService(stress)
	log.Info("Creating metrics service.", "MetricsService.Namespace", metricsService.Namespace,
		"MetricsService.Name", metricsService.Name)
	err := client.Create(context.TODO(), metricsService)
	if err != nil {
		log.Error(err, "Failed to create metrics service.", "MetricsService.Namespace",
			metricsService.Namespace, "MetricsService.Name", metricsService.Name)
		return reconcile.Result{}, err
	}
	return reconcile.Result{Requeue: true}, nil
}

func GetMetricsServiceName(stress *api.Stress) string {
	return fmt.Sprintf("%s-metrics", stress.Name)
}

func newMetricsService(stress *api.Stress) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      GetMetricsServiceName(stress),
			Namespace: stress.Namespace,
			Labels:    tlp.LabelsForStress(stress.Name),
			OwnerReferences: []metav1.OwnerReference{
				stress.CreateOwnerReference(),
			},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Port: 9500,
					Name: metricsPort,
					Protocol: corev1.ProtocolTCP,
					TargetPort: intstr.IntOrString{
						Type: intstr.Int,
						IntVal: 9500,
					},
				},
			},
			Selector: tlp.LabelsForStress(stress.Name),
		},
	}
}

func ServiceMonitorKindExists() (bool, error) {
	return discoveryClient.KindExists(prometheus.SchemeGroupVersion.String(), ServiceMonitorKind)
}

func PrometheusKindExists() (bool, error) {
	return discoveryClient.KindExists(prometheus.SchemeGroupVersion.String(), PrometheusKind)
}

func GetPrometheus(namespace string, client client.Client) (*prometheus.Prometheus, error) {
	instance := &prometheus.Prometheus{}
	err := client.Get(context.TODO(), types.NamespacedName{Namespace: namespace, Name: PrometheusName}, instance)

	return instance, err
}

func CreatePrometheus(namespace string, client client.Client, log logr.Logger) (reconcile.Result, error) {
	if err := k8s.CreateServiceAccount(client, namespace, PrometheusName); err != nil {
		return reconcile.Result{}, fmt.Errorf("failed to create prometheus service account: %s", err)
	}
	if err := createPrometheusRole(client, namespace); err != nil {
		return reconcile.Result{}, fmt.Errorf("failed to create prometheus role: %s", err)
	}
	if err := createPrometheusRoleBinding(client, namespace); err != nil {
		return reconcile.Result{}, fmt.Errorf("failed to create prometheus role binding: %s", err)
	}
	// TODO create service to expose prometheus UI
	instance := newPrometheus(namespace)
	log.Info("Creating Prometheus", "Prometheus.Namespace", instance.Namespace, "Prometheus.Name",
		instance.Name)
	if err := k8s.CreateResource(client, instance); err != nil {
		log.Error(err, "Failed to create Prometheus", "Prometheus.Namespace", instance.Namespace,
			"Prometheus.Name", instance.Name)
		return reconcile.Result{}, err
	}
	return reconcile.Result{Requeue: true, RequeueAfter: 10 * time.Second}, nil
}

func createPrometheusRole(client client.Client, namespace string) error {
	role := &rbac.Role{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      PrometheusName,
		},
		Rules: []rbac.PolicyRule{
			{
				APIGroups: []string { "" },
				Verbs: []string {"get", "list", "watch"},
				Resources: []string {"nodes", "services", "endpoints", "pods"},
			},
		},
	}
	return k8s.CreateResource(client, role)
}

func createPrometheusRoleBinding(client client.Client, namespace string) error {
	roleBinding := &rbac.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      PrometheusName,
		},
		RoleRef: rbac.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "Role",
			Name:     PrometheusName,
		},
		Subjects: []rbac.Subject{
			{
				Namespace: namespace,
				Name:      PrometheusName,
				Kind:      "ServiceAccount",
			},
		},
	}
	return k8s.CreateResource(client, roleBinding)
}

func newPrometheus(namespace string) *prometheus.Prometheus {
	selector := metav1.LabelSelector{
		MatchExpressions: []metav1.LabelSelectorRequirement{
			{
				Key: "app",
				Operator: metav1.LabelSelectorOpIn,
				Values: []string{"stress"},
			},
		},
	}

		return &prometheus.Prometheus{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      PrometheusName,
		},
		Spec: prometheus.PrometheusSpec{
			ServiceAccountName:     PrometheusName,
			ServiceMonitorSelector: &selector,
			Resources: v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceMemory: resource.MustParse("400Mi"),
				},
			},
			EnableAdminAPI: false,
		},
	}
}

func GetServiceMonitor(namespace string, client client.Client) (*prometheus.ServiceMonitor, error) {
	serviceMonitor := &prometheus.ServiceMonitor{}
	err := client.Get(context.TODO(), types.NamespacedName{Namespace: namespace, Name: ServiceMonitorName}, serviceMonitor)

	return serviceMonitor, err
}

func CreateServiceMonitor(namespace string, client client.Client, log logr.Logger) (reconcile.Result, error) {
	serviceMonitor := newServiceMonitor(namespace)
	log.Info("Creating service monitor", "ServiceMonitor.Namespace", serviceMonitor.Namespace,
		"ServiceMonitor.Name", serviceMonitor.Name)
	err := client.Create(context.TODO(), serviceMonitor)
	if err != nil {
		log.Error(err, "Failed to create service monitor", "ServiceMonitor.Namespace",
			serviceMonitor.Namespace, "ServiceMonitor.Name", serviceMonitor.Name)
		return reconcile.Result{}, err
	}
	return reconcile.Result{Requeue: true}, nil
}

func newServiceMonitor(namespace string) *prometheus.ServiceMonitor {
	return &prometheus.ServiceMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name: ServiceMonitorName,
			Namespace: namespace,
			Labels: map[string]string{
				"app": "stress",
			},
		},
		Spec: prometheus.ServiceMonitorSpec{
			Selector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "stress",
				},
			},
			Endpoints: []prometheus.Endpoint{
				{
					Port: metricsPort,
				},
			},
 		},
	}
}

func GetPrometheusService(namespace string, client client.Client) (*corev1.Service, error) {
	metricsService := &corev1.Service{}
	err := client.Get(context.TODO(), types.NamespacedName{Namespace: namespace, Name: PrometheusName}, metricsService)

	return metricsService, err
}

func CreatePrometheusService(namespace string, client client.Client, log logr.Logger) (reconcile.Result, error) {
	metricsService := newPrometheusService(namespace)
	log.Info("Creating Prometheus service.", "PrometheusService.Namespace", namespace,
		"PrometheusService.Name", PrometheusName)
	err := client.Create(context.TODO(), metricsService)
	if err != nil {
		log.Error(err, "Failed to create Prometheus service.", "PrometheusService.Namespace", namespace,
			"PrometheusService.Name", PrometheusName)
		return reconcile.Result{}, err
	}
	return reconcile.Result{Requeue: true}, nil
}

func newPrometheusService(namespace string) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      PrometheusName,
			Namespace: namespace,
			Labels: map[string]string{
				"app": "prometheus",
			},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Port: 9090,
					Name: "web",
					Protocol: corev1.ProtocolTCP,
					TargetPort: intstr.IntOrString{
						StrVal: "web",
					},
				},
			},
			Selector: map[string]string{
				"prometheus": PrometheusName,
			},
		},
	}
}