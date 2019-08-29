package e2e

import (
	goctx "context"
	casskop "github.com/Orange-OpenSource/cassandra-k8s-operator/pkg/apis/db/v1alpha1"
	"github.com/jsanda/tlp-stress-operator/pkg/apis"
	tlp "github.com/jsanda/tlp-stress-operator/pkg/apis/thelastpickle/v1alpha1"
	"github.com/jsanda/tlp-stress-operator/test/e2eutil"
	framework "github.com/operator-framework/operator-sdk/pkg/test"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"testing"
	"time"
	"github.com/jsanda/tlp-stress-operator/pkg/apis/thelastpickle/v1alpha1"
)

var (
	retryInterval        = time.Second * 5
	timeout              = time.Second * 60
	cleanupRetryInterval = time.Second * 1
	cleanupTimeout       = time.Second * 5
)

func noCleanup() *framework.CleanupOptions {
	return &framework.CleanupOptions{}
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
	err = e2eutil.WaitForOperatorDeployment(t, f, namespace, "tlp-stress-operator", retryInterval, timeout)
	if err != nil {
		t.Fatalf("Failed waiting for tlp-stress operator deployment: %s\n", err)
	}

	if err = createCassandraCluster(t, f, ctx); err != nil {
		t.Fatalf("Failed to create CassandraCluster: %s", err)
	}

	if err = createTLPStress(t, f); err != nil {
		t.Fatalf("Failed to create TLPStress: %s", err)
	}

	if err = e2eutil.WaitForCassKopCluster(t, f, namespace, "tlp-stress-test", 10 * time.Second, 3 * time.Minute); err != nil {
		t.Fatalf("Failed waiting for CassandraCluster to become ready: %s\n", err)
	}

	if err = e2eutil.WaitForTLPStressToStart(t, f, namespace, "tlp-stress-test", 10 * time.Second, 60 * time.Second); err != nil {
		t.Errorf("Failed waiting for TLPStress to start: %s\n", err)
	}

	if err = e2eutil.WaitForTLPStressToFinish(t, f, namespace, "tlp-stress-test", 10 * time.Second, 60 * time.Second); err != nil {
		t.Errorf("Failed waiting for TLPStress to finish: %s\n", err)
	}
}

func createTLPStress(t *testing.T, f *framework.Framework) error {
	tlpStress := v1alpha1.TLPStress{
		ObjectMeta: metav1.ObjectMeta{
			Name: "tlp-stress-test",
			Namespace: f.Namespace,
		},
		Spec: v1alpha1.TLPStressSpec{
			StressConfig: v1alpha1.TLPStressConfig{
				Workload: v1alpha1.KeyValueWorkload,
				Duration: "20s",
			},
			CassandraConfig: v1alpha1.CassandraConfig{
				CassandraCluster:&v1alpha1.CassandraCluster{
					Name: "tlp-stress-test",
				},
			},
		},
	}
	return f.Client.Create(goctx.TODO(), &tlpStress, noCleanup())
}

func createCassandraCluster(t *testing.T, f *framework.Framework, ctx *framework.TestCtx) error {
	cc := casskop.CassandraCluster{
		TypeMeta:   metav1.TypeMeta{
			Kind: "CassandraCluster",
			APIVersion: "db.orange.com/v1alpha",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "tlp-stress-test",
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