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
	//return &framework.CleanupOptions{}
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
// tlp-stress-operator. When each of the subtest runs, there will already be a cluster
// available.
func TestTLPStressWithExistingCluster(t *testing.T) {
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
	ctx, f := e2eutil.InitOperator(t)
	defer ctx.Cleanup()

	if err := createCassandraCluster(cassandraClusterName, t, f, ctx); err != nil {
		t.Fatalf("Failed to create CassandraCluster: %s", err)
	}

	if err := e2eutil.WaitForCassKopCluster(t, f, f.Namespace, cassandraClusterName, 10 * time.Second, 3 * time.Minute); err != nil {
		t.Fatalf("Failed waiting for CassandraCluster to become ready: %s\n", err)
	}

	// run subtests
	t.Run("tlpstress-group", func(t *testing.T) {
		t.Run("RunOneTLPStress", e2eTest(runOneTLPStress, t, f, ctx))
		t.Run("RunTwoTLPStress", e2eTest(runTwoTLPStress, t, f, ctx))
	})
}

func runOneTLPStress(t *testing.T, f *framework.Framework, ctx *framework.TestCtx) {
	namespace := f.Namespace
	name := "tlpstress-test"

	if err := createTLPStress(name, t, f, ctx); err != nil {
		t.Fatalf("Failed to create TLPStress: %s", err)
	}

	if err := e2eutil.WaitForTLPStressToStart(t, f, namespace, name, 10 * time.Second, 1 * time.Minute); err != nil {
		t.Errorf("Failed waiting for TLPStress to start: %s\n", err)
	}

	if err := e2eutil.WaitForTLPStressToFinish(t, f, namespace, name, 1, 10 * time.Second, 3 * time.Minute); err != nil {
		t.Errorf("Failed waiting for TLPStress to finish: %s\n", err)
	}

	tlpStress := &v1alpha1.TLPStress{}
	if err := f.Client.Get(goctx.TODO(), types.NamespacedName{Namespace: namespace, Name: name}, tlpStress); err != nil {
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

func runTwoTLPStress(t *testing.T,  f *framework.Framework, ctx *framework.TestCtx) {
	name := "tlpstress-test-two"

	tlpStress := &v1alpha1.TLPStress{
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
					Name: cassandraClusterName,
				},
			},
			JobConfig: v1alpha1.JobConfig{
				Parallelism: int32Ptr(2),
			},
		},
	}
	if err := f.Client.Create(goctx.TODO(), tlpStress, cleanupWithPolling(ctx)); err != nil {
		t.Fatalf("Failed to create TLPStress (%s): %s", name, err)
	}

	if err := e2eutil.WaitForTLPStressToStart(t, f, f.Namespace, name, 10 * time.Second, 1 * time.Minute); err != nil {
		t.Errorf("Failed waiting for TLPStress (%s) to start: %s\n", name, err)
	}

	if err := e2eutil.WaitForTLPStressToFinish(t, f, f.Namespace, name, 2, 10 * time.Second, 3 * time.Minute); err != nil {
		t.Errorf("Failed waiting for TLPStress (%s) to finish: %s\n", name, err)
	}

	tlpStress = &v1alpha1.TLPStress{}
	if err := f.Client.Get(goctx.TODO(), types.NamespacedName{Namespace: f.Namespace, Name: name}, tlpStress); err != nil {
		t.Fatalf("Failed to get TLPStress instance (%s): %s", name, err)
	}

	jobStatus := tlpStress.Status.JobStatus
	if jobStatus == nil {
		t.Fatal("job status should not be nil")
	}

	if jobStatus.Succeeded != 2 || jobStatus.Failed != 0 {
		t.Fatalf("Expected succeeded(2) and failed(0) but got succeeded(%d) and failed(%d)", jobStatus.Succeeded,
			jobStatus.Failed)
	}
}

func createTLPStress(name string, t *testing.T, f *framework.Framework, ctx *framework.TestCtx) error {
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
					Name: cassandraClusterName,
				},
			},
		},
	}
	return f.Client.Create(goctx.TODO(), &tlpStress, cleanupWithPolling(ctx))
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
	return f.Client.Create(goctx.TODO(), &cc, cleanupWithPolling(ctx))
}

func stringPtr(s string) *string {
	return &s
}

func int32Ptr(n int32) *int32 {
	return &n
}