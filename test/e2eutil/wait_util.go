package e2eutil

import (
	goctx "context"
	casskop "github.com/Orange-OpenSource/cassandra-k8s-operator/pkg/apis/db/v1alpha1"
	"github.com/jsanda/stress-operator/pkg/apis/thelastpickle/v1alpha1"
	framework "github.com/operator-framework/operator-sdk/pkg/test"
	"github.com/operator-framework/operator-sdk/pkg/test/e2eutil"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"testing"
	"time"
)

// Waits for the CassandraCluster with namespace/name to be ready. Specifically, this
// function checks for CassandraCluster.State.Phase == Running.
func WaitForCassKopCluster(
	t *testing.T,
	f *framework.Framework,
	namespace string,
	name string,
	retryInterval time.Duration,
	timeout time.Duration,) error {

	return wait.Poll(retryInterval, timeout, func() (bool, error) {
		cc := &casskop.CassandraCluster{}
		err := f.Client.Get(goctx.TODO(), types.NamespacedName{Name: name, Namespace: namespace}, cc)
		if err != nil {
			if apierrors.IsNotFound(err) {
				t.Logf("Waiting for availability of CassandraCluster %s\n", name)
				return false, nil
			}
			return false, err
		}
		if cc.Status.Phase != "Running" {
			t.Logf("Waiting for CassandraCassandra %s (%s)\n", name, cc.Status.Phase)
			return false, nil
		}
		return true, nil
	})
}

// Waits for the Stress instance specified by namespace/name to start. Specifically this
// functions blocks until Stress.Status.JobStatus.Active > 0.
func WaitForStressToStart(t *testing.T,
	f *framework.Framework,
	namespace string,
	name string,
	retryInterval time.Duration,
	timeout time.Duration,) error {

	return wait.Poll(retryInterval, timeout, func() (bool, error) {
		stress := &v1alpha1.Stress{}
		err := f.Client.Get(goctx.TODO(), types.NamespacedName{Namespace: namespace, Name: name}, stress)
		if err != nil {
			if apierrors.IsNotFound(err) {
				t.Logf("Waiting for availability of Stress %s\n", name)
				return false, nil
			}
			t.Logf("Failed to get Stress (%s): %s", name, err)
			return false, err
		}
		if stress.Status.JobStatus == nil {
			t.Log("Stress.Status.JobStatus is nil")
			return false, nil
		}
		if stress.Status.JobStatus.Active == 0 {
			t.Logf("Waiting for Stress %s to start\n", name)
			return false, nil
		}
		return true, nil
	})
}

// Waits for the Stress instance specified by namespace/name to finish. Specifically this
// function blocks until the succeeded + failed job runs equals completions.
func WaitForStressToFinish(t *testing.T,
	f *framework.Framework,
	namespace string,
	name string,
	completions int32,
	retryInterval time.Duration,
	timeout time.Duration,) error {

	return wait.Poll(retryInterval, timeout, func() (bool, error) {
		stress := &v1alpha1.Stress{}
		err := f.Client.Get(goctx.TODO(), types.NamespacedName{Namespace: namespace, Name: name}, stress)
		if err != nil {
			if apierrors.IsNotFound(err) {
				t.Logf("Waiting for availability of Stress %s\n", name)
				return false, nil
			}
			return false, err
		}
		if stress.Status.JobStatus == nil {
			t.Log("Stress.Status.JobStatus is nil")
			return false, nil
		}

		if (stress.Status.JobStatus.Succeeded + stress.Status.JobStatus.Failed) != completions {
			t.Logf("Waiting for Stress (%s) to complete (%d). There are: succeeded(%d), failed(%d)\n", name,
				completions, stress.Status.JobStatus.Succeeded, stress.Status.JobStatus.Failed)
			return false, nil
		}
		return true, nil
	})
}

func WaitForOperatorDeployment(t *testing.T,
	f *framework.Framework,
	namespace string,
	name string,
	retryInterval time.Duration,
	timeout time.Duration,) error {

	return e2eutil.WaitForDeployment(t, f.KubeClient, namespace, name, 1, retryInterval, timeout)
}
