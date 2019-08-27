package e2e

import (
	"github.com/jsanda/tlp-stress-operator/pkg/apis"
	"github.com/operator-framework/operator-sdk/pkg/test/e2eutil"
	"k8s.io/kubernetes/pkg/apis/apps"
	"testing"
	"time"
	tlp "github.com/jsanda/tlp-stress-operator/pkg/apis/thelastpickle/v1alpha1"
	framework "github.com/operator-framework/operator-sdk/pkg/test"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	appsv1 "k8s.io/api/apps/v1"
)

var (
	retryInterval        = time.Second * 5
	timeout              = time.Second * 60
	cleanupRetryInterval = time.Second * 1
	cleanupTimeout       = time.Second * 5
)

func TestTLPStress(t *testing.T) {
	tlpStressList := &tlp.TLPStressList{}
	err := framework.AddToFrameworkScheme(apis.AddToScheme, tlpStressList)
	if err != nil {
		t.Fatalf("failed to add custom resource scheme to framework: %v", err)
	}
	// run subtests
	t.Run("tlpstress-group", func(t *testing.T) {
		t.Run("RunOneTLPStress", runOneTLPStress)
	})
}

func runOneTLPStress(t *testing.T) {
	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup()

	err := ctx.InitializeClusterResources(&framework.CleanupOptions{TestContext: ctx, Timeout: cleanupTimeout, RetryInterval: cleanupRetryInterval})
	if err != nil {
		t.Fatalf("failed to initialize cluster resources: %v", err)
	}
	t.Log("Initialized cluster resources")
	namespace, err := ctx.GetNamespace()
	if err != nil {
		t.Fatal(err)
	}
	// get global framework variables
	f := framework.Global
	// wait for tlp-stress-operator to be ready
	err = e2eutil.WaitForDeployment(t, f.KubeClient, namespace, "tlp-stress-operator", 1, retryInterval, timeout)
	if err != nil {
		t.Fatal(err)
	}
}

func createCassandraCluster(t *testing.T, f *framework.Framework) {
	createHeadlessService(t, f)
	createCassandraCluster(t, f)
}

func createHeadlessService(t *testing.T, f *framework.Framework) {
	_, err := f.KubeClient.CoreV1().Services(f.Namespace).Create(&corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: "cassandra",
			Namespace: f.Namespace,
			Labels: map[string]string{
				"app": "cassandra",
			},
		},
		Spec: corev1.ServiceSpec{
			ClusterIP: "None",
			Ports: []corev1.ServicePort{
				{Port: 9042},
			},
			Selector: map[string]string{
				"app": "cassandra",
			},
		},
	})
	if err != nil {
		t.Fatalf("Failed to create Cassandra service: %s", err)
	}
}

func createStatefulSet(t *testing.T, f *framework.Framework) {
	_, err := f.KubeClient.AppsV1().StatefulSets(f.Namespace).Create(&appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name: "cassandra",
			Namespace: f.Namespace,
			Labels: map[string]string{
				"app": "cassandra",
			},
		},
		Spec: appsv1.StatefulSetSpec{
			ServiceName: "cassandra",
			Replicas: int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchExpressions: []metav1.LabelSelectorRequirement{
					{
						Key:      "app",
						Operator: metav1.LabelSelectorOpIn,
						Values:   []string{"cassandra"},
					},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("Failed to create Cassandra statefulset: %s", err)
	}
}

func int32Ptr(n int32) *int32 {
	return &n
}