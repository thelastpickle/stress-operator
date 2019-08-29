package e2e

import (
	goctx "context"
	casskop "github.com/Orange-OpenSource/cassandra-k8s-operator/pkg/apis/db/v1alpha1"
	"github.com/jsanda/tlp-stress-operator/pkg/apis"
	"github.com/jsanda/tlp-stress-operator/pkg/apis/thelastpickle/v1alpha1"
	tlp "github.com/jsanda/tlp-stress-operator/pkg/apis/thelastpickle/v1alpha1"
	"github.com/jsanda/tlp-stress-operator/test/e2eutil"
	framework "github.com/operator-framework/operator-sdk/pkg/test"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"testing"
	"time"
	"k8s.io/apimachinery/pkg/types"
)

var (
	retryInterval        = time.Second * 5
	timeout              = time.Second * 60
	cleanupRetryInterval = time.Second * 1
	cleanupTimeout       = time.Second * 5
)

func noCleanup() *framework.CleanupOptions {
	//return &framework.CleanupOptions{}
	return nil
}

func TestTLPStress(t *testing.T) {
	tlpStressList := &tlp.TLPStressList{}
	err := framework.AddToFrameworkScheme(apis.AddToScheme, tlpStressList)
	if err != nil {
		t.Fatalf("failed to add custom resource scheme to framework: %v", err)
	}
	framework.Global.Scheme.AddKnownTypes(schema.GroupVersion{Group: "db.orange.com", Version: "v1alpha1"},
		&casskop.CassandraCluster{},
		&casskop.CassandraClusterList{},
		&metav1.ListOptions{},
	)
	// run subtests
	t.Run("tlpstress-group", func(t *testing.T) {
		t.Run("RunOneTLPStress", runOneTLPStress)
	})
}

func runOneTLPStress(t *testing.T) {
	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup()

	//err := ctx.InitializeClusterResources(&framework.CleanupOptions{TestContext: ctx, Timeout: cleanupTimeout, RetryInterval: cleanupRetryInterval})
	err := ctx.InitializeClusterResources(noCleanup())
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
	err = e2eutil.WaitForOperatorDeployment(t, f, namespace, "tlp-stress-operator", retryInterval, timeout)
	if err != nil {
		t.Fatalf("Failed waiting for tlp-stress operator deployment: %s\n", err)
	}

	name := "tlpstress-test"

	if err = createCassandraCluster(name, t, f, ctx); err != nil {
		t.Fatalf("Failed to create CassandraCluster: %s", err)
	}

	if err = e2eutil.WaitForCassKopCluster(t, f, namespace, name, 10 * time.Second, 3 * time.Minute); err != nil {
		t.Fatalf("Failed waiting for CassandraCluster to become ready: %s\n", err)
	}

	if err = createTLPStress(name, t, f); err != nil {
		t.Fatalf("Failed to create TLPStress: %s", err)
	}

	if err = e2eutil.WaitForTLPStressToStart(t, f, namespace, name, 10 * time.Second, 1 * time.Minute); err != nil {
		t.Errorf("Failed waiting for TLPStress to start: %s\n", err)
	}

	if err = e2eutil.WaitForTLPStressToFinish(t, f, namespace, name, 10 * time.Second, 3 * time.Minute); err != nil {
		t.Errorf("Failed waiting for TLPStress to finish: %s\n", err)
	}

	tlpStress := &v1alpha1.TLPStress{}
	if err = f.Client.Get(goctx.TODO(), types.NamespacedName{Namespace: namespace, Name: name}, tlpStress); err != nil {
		t.Fatal("Failed to get TLPStress instance")
	}

	jobStatus := tlpStress.Status.JobStatus
	if jobStatus == nil {
		t.Fatal("job status should not be nil")
	}

	if jobStatus.Succeeded != 1 || jobStatus.Failed != 0 {
		t.Fatalf("Expected succeeded(1) and failed(0) but got succeeded(%d) and failed(%d)", jobStatus.Succeeded,
			jobStatus.Failed)
	}
}

func createTLPStress(name string, t *testing.T, f *framework.Framework) error {
	tlpStress := v1alpha1.TLPStress{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Namespace: f.Namespace,
		},
		Spec: v1alpha1.TLPStressSpec{
			StressConfig: v1alpha1.TLPStressConfig{
				Workload: v1alpha1.KeyValueWorkload,
				Iterations: stringPtr("500"),
			},
			CassandraConfig: v1alpha1.CassandraConfig{
				CassandraCluster:&v1alpha1.CassandraCluster{
					Name: name,
				},
			},
		},
	}
	return f.Client.Create(goctx.TODO(), &tlpStress, noCleanup())
}

func createCassandraCluster(name string, t *testing.T, f *framework.Framework, ctx *framework.TestCtx) error {
	cc := casskop.CassandraCluster{
		TypeMeta:   metav1.TypeMeta{
			Kind: "CassandraCluster",
			APIVersion: "db.orange.com/v1alpha",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Namespace: f.Namespace,
		},
		Spec: casskop.CassandraClusterSpec{
			DeletePVC: true,
			DataCapacity: "5Gi",
			Resources: casskop.CassandraResources{
				Requests: casskop.CPUAndMem{
					CPU: "500m",
					Memory: "1Gi",
				},
				Limits: casskop.CPUAndMem{
					CPU: "500m",
					Memory: "1Gi",
				},
			},
		},
	}
	return f.Client.Create(goctx.TODO(), &cc, noCleanup())
}

func stringPtr(s string) *string {
	return &s
}