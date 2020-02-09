package e2e

import (
	goctx "context"
	casskop "github.com/Orange-OpenSource/cassandra-k8s-operator/pkg/apis/db/v1alpha1"
	casskopapi "github.com/Orange-OpenSource/cassandra-k8s-operator/pkg/apis"
	"github.com/thelastpickle/stress-operator/pkg/apis"
	"github.com/thelastpickle/stress-operator/pkg/apis/thelastpickle/v1alpha1"
	tlp "github.com/thelastpickle/stress-operator/pkg/apis/thelastpickle/v1alpha1"
	"github.com/thelastpickle/stress-operator/test/e2eutil"
	framework "github.com/operator-framework/operator-sdk/pkg/test"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"testing"
	"time"
)

var (
	retryInterval        = time.Second * 5
	timeout              = time.Second * 60
	cleanupRetryInterval = time.Second * 1
	cleanupTimeout       = time.Second * 60
)

const (
	cassandraClusterName = "tlpstress-cluster"
)

func noCleanup() *framework.CleanupOptions {
	return nil
}

func cleanupWithPolling(ctx *framework.TestCtx) *framework.CleanupOptions {
	return &framework.CleanupOptions{
		TestContext:   ctx,
		Timeout:       cleanupTimeout,
		RetryInterval: cleanupRetryInterval,
	}
}

type TestFunc func(t *testing.T, f *framework.Framework, ctx *framework.TestCtx)

func e2eTest(fn TestFunc, t *testing.T, f *framework.Framework, ctx *framework.TestCtx) func(t *testing.T) {
	return func(t *testing.T) {
		fn(t, f, ctx)
	}
}

// This test actually does create a Cassandra cluster, but it does so independent of the
// stress-operator. When each of the subtest runs, there will already be a cluster
// available.
func TestStressWithExistingCluster(t *testing.T) {
	stressList := &tlp.StressList{}
	err := framework.AddToFrameworkScheme(apis.AddToScheme, stressList)
	if err != nil {
		t.Fatalf("failed to add custom resource scheme to framework: %v", err)
	}
	ccList := &casskop.CassandraClusterList{}
	if err = framework.AddToFrameworkScheme(casskopapi.AddToScheme, ccList); err != nil {
		t.Fatalf("failed to add custom resource scheme to framework: %v", err)
	}
	ctx, f := e2eutil.InitOperator(t)
	defer ctx.Cleanup()

	namespace, err := ctx.GetNamespace()
	if err != nil {
		t.Fatalf("Failed to get namespace: %s", err)
	}

	if err := createCassandraCluster(cassandraClusterName, namespace, f, ctx); err != nil {
		t.Fatalf("Failed to create CassandraCluster: %s", err)
	}

	if err := e2eutil.WaitForCassKopCluster(t, f, namespace, cassandraClusterName, 10 * time.Second, 3 * time.Minute); err != nil {
		t.Fatalf("Failed waiting for CassandraCluster to become ready: %s\n", err)
	}

	// run subtests
	t.Run("stress-group", func(t *testing.T) {
		t.Run("RunOneStress", e2eTest(runOneStress, t, f, ctx))
		t.Run("RunTwoStress", e2eTest(runTwoStress, t, f, ctx))
	})
}

func runOneStress(t *testing.T, f *framework.Framework, ctx *framework.TestCtx) {
	namespace, err := ctx.GetNamespace()
	if err != nil {
		t.Fatalf("Failed to get namespace: %s", err)
	}
	name := "stress-test"

	if err := createStress(name, namespace, f, ctx); err != nil {
		t.Fatalf("Failed to create Stress: %s", err)
	}

	if err := e2eutil.WaitForStressToStart(t, f, namespace, name, 10 * time.Second, 1 * time.Minute); err != nil {
		t.Errorf("Failed waiting for Stress to start: %s\n", err)
	}

	if err := e2eutil.WaitForStressToFinish(t, f, namespace, name, 1, 10 * time.Second, 3 * time.Minute); err != nil {
		t.Errorf("Failed waiting for Stress to finish: %s\n", err)
	}

	stress := &v1alpha1.Stress{}
	if err := f.Client.Get(goctx.TODO(), types.NamespacedName{Namespace: namespace, Name: name}, stress); err != nil {
		t.Fatal("Failed to get Stress instance")
	}

	jobStatus := stress.Status.JobStatus
	if jobStatus == nil {
		t.Fatal("job status should not be nil")
	}

	if jobStatus.Succeeded != 1 || jobStatus.Failed != 0 {
		t.Fatalf("Expected succeeded(1) and failed(0) but got succeeded(%d) and failed(%d)", jobStatus.Succeeded,
			jobStatus.Failed)
	}
}

func runTwoStress(t *testing.T,  f *framework.Framework, ctx *framework.TestCtx) {
	name := "stress-test-two"

	namespace, err := ctx.GetNamespace()
	if err != nil {
		t.Fatalf("Failed to get namespace: %s", err)
	}

	stress := &v1alpha1.Stress{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Namespace: namespace,
		},
		Spec: v1alpha1.StressSpec{
			StressConfig: v1alpha1.StressConfig{
				Workload: v1alpha1.KeyValueWorkload,
				Iterations: stringPtr("50"),
			},
			CassandraConfig: v1alpha1.CassandraConfig{
				CassandraCluster:&v1alpha1.CassandraCluster{
					Name: cassandraClusterName,
				},
			},
			JobConfig: v1alpha1.JobConfig{
				Parallelism: int32Ptr(2),
			},
		},
	}
	if err := f.Client.Create(goctx.TODO(), stress, cleanupWithPolling(ctx)); err != nil {
		t.Fatalf("Failed to create Stress (%s): %s", name, err)
	}

	if err := e2eutil.WaitForStressToStart(t, f, namespace, name, 10 * time.Second, 1 * time.Minute); err != nil {
		t.Errorf("Failed waiting for Stress (%s) to start: %s\n", name, err)
	}

	if err := e2eutil.WaitForStressToFinish(t, f, namespace, name, 2, 1 * time.Second, 3 * time.Minute); err != nil {
		t.Errorf("Failed waiting for Stress (%s) to finish: %s\n", name, err)
	}

	stress = &v1alpha1.Stress{}
	if err := f.Client.Get(goctx.TODO(), types.NamespacedName{Namespace: namespace, Name: name}, stress); err != nil {
		t.Fatalf("Failed to get Stress instance (%s): %s", name, err)
	}

	jobStatus := stress.Status.JobStatus
	if jobStatus == nil {
		t.Fatal("job status should not be nil")
	}

	if jobStatus.Succeeded != 2 || jobStatus.Failed != 0 {
		t.Fatalf("Expected succeeded(2) and failed(0) but got succeeded(%d) and failed(%d)", jobStatus.Succeeded,
			jobStatus.Failed)
	}
}

func createStress(name string, namespace string, f *framework.Framework, ctx *framework.TestCtx) error {
	stress := v1alpha1.Stress{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Namespace: namespace,
		},
		Spec: v1alpha1.StressSpec{
			StressConfig: v1alpha1.StressConfig{
				Workload: v1alpha1.KeyValueWorkload,
				Iterations: stringPtr("50"),
			},
			CassandraConfig: v1alpha1.CassandraConfig{
				CassandraCluster:&v1alpha1.CassandraCluster{
					Name: cassandraClusterName,
				},
			},
		},
	}
	return f.Client.Create(goctx.TODO(), &stress, cleanupWithPolling(ctx))
}

func createCassandraCluster(name string, namespace string, f *framework.Framework, ctx *framework.TestCtx) error {
	cc := casskop.CassandraCluster{
		TypeMeta:   metav1.TypeMeta{
			Kind: "CassandraCluster",
			APIVersion: "db.orange.com/v1alpha",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Namespace: namespace,
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
	return f.Client.Create(goctx.TODO(), &cc, cleanupWithPolling(ctx))
}

func stringPtr(s string) *string {
	return &s
}

func int32Ptr(n int32) *int32 {
	return &n
}