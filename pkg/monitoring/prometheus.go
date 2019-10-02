package monitoring

import (
	"context"
	"fmt"
	prometheus "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/go-logr/logr"
	api "github.com/jsanda/tlp-stress-operator/pkg/apis/thelastpickle/v1alpha1"
	tlp "github.com/jsanda/tlp-stress-operator/pkg/tlpstress"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func GetMetricsService(tlpStress *api.TLPStress, client client.Client) (*corev1.Service, error) {
	metricsService := &corev1.Service{}
	err := client.Get(context.TODO(), types.NamespacedName{Namespace: tlpStress.Namespace, Name: GetMetricsServiceName(tlpStress)}, metricsService)

	return metricsService, err
}

func CreateMetricsService(tlpStress *api.TLPStress, client client.Client, log logr.Logger) (reconcile.Result, error) {
	metricsService := newMetricsService(tlpStress)
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

func GetMetricsServiceName(tlpStress *api.TLPStress) string {
	return fmt.Sprintf("%s-metrics", tlpStress.Name)
}

func newMetricsService(tlpStress *api.TLPStress) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      GetMetricsServiceName(tlpStress),
			Namespace: tlpStress.Namespace,
			Labels:    tlp.LabelsForTLPStress(tlpStress.Name),
			OwnerReferences: []metav1.OwnerReference{
				tlpStress.CreateOwnerReference(),
			},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Port: 9500,
					Name: "metrics",
					Protocol: corev1.ProtocolTCP,
					TargetPort: intstr.IntOrString{
						Type: intstr.Int,
						IntVal: 9500,
					},
				},
			},
			Selector: tlp.LabelsForTLPStress(tlpStress.Name),
		},
	}
}

func GetServiceMonitor(tlpStress *api.TLPStress, client client.Client) (*prometheus.ServiceMonitor, error) {
	metricsSvc := GetMetricsServiceName(tlpStress)
	serviceMonitor := &prometheus.ServiceMonitor{}
	err := client.Get(context.TODO(), types.NamespacedName{Namespace: tlpStress.Namespace, Name: metricsSvc}, serviceMonitor)

	return serviceMonitor, err
}

func CreateServiceMonitor(svc *corev1.Service, client client.Client, log logr.Logger) (reconcile.Result, error) {
	serviceMonitor := newServiceMonitor(svc)
	log.Info("Creating service monitor.", "ServiceMonitor.Namespace", serviceMonitor.Namespace,
		"ServiceMonitor.Name", serviceMonitor.Name)
	err := client.Create(context.TODO(), serviceMonitor)
	if err != nil {
		log.Error(err, "Failed to create service monitor", "ServiceMonitor.Namespace",
			serviceMonitor.Namespace, "ServiceMonitor.Name", serviceMonitor.Name)
		return reconcile.Result{}, err
	}
	return reconcile.Result{Requeue: true}, nil
}

func newServiceMonitor(svc *corev1.Service) *prometheus.ServiceMonitor {
	return &prometheus.ServiceMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name: svc.Name,
			Namespace: svc.Namespace,
			Labels: svc.Labels,
			OwnerReferences: svc.OwnerReferences,
		},
		Spec: prometheus.ServiceMonitorSpec{
			Selector: metav1.LabelSelector{
				MatchLabels: svc.Spec.Selector,
			},
			Endpoints: getEndpoints(svc),
		},
	}
}

func getEndpoints(s *corev1.Service) []prometheus.Endpoint {
	var endpoints []prometheus.Endpoint
	for _, port := range s.Spec.Ports {
		endpoints = append(endpoints, prometheus.Endpoint{Port: port.Name})
	}
	return endpoints
}