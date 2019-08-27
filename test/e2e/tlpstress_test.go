package e2e

import (
	"github.com/jsanda/tlp-stress-operator/pkg/apis"
	"github.com/operator-framework/operator-sdk/pkg/test/e2eutil"
	"testing"
	"time"
	tlp "github.com/jsanda/tlp-stress-operator/pkg/apis/thelastpickle/v1alpha1"
	framework "github.com/operator-framework/operator-sdk/pkg/test"
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
	// wait for memcached-operator to be ready
	err = e2eutil.WaitForDeployment(t, f.KubeClient, namespace, "tlp-stress-operator", 1, retryInterval, timeout)
	if err != nil {
		t.Fatal(err)
	}


}