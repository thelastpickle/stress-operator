package e2eutil

import (
	goctx "context"
	casskop "github.com/Orange-OpenSource/cassandra-k8s-operator/pkg/apis/db/v1alpha1"
	"github.com/jsanda/tlp-stress-operator/pkg/apis/thelastpickle/v1alpha1"
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
		if (cc.Status.Phase != "Running") {
			t.Logf("Waiting for CassandraCassandra %s (%s)\n", name, cc.Status.Phase)
			return false, nil
		}
		return true, nil
	})
}

// Waits for the TLPStress instance specified by namespace/name to start. Specifically this
// functions blocks until TLPStress.Status.JobStatus.Active > 0.
func WaitForTLPStressToStart(t *testing.T,
	f *framework.Framework,
	namespace string,
	name string,
	retryInterval time.Duration,
	timeout time.Duration,) error {

	return wait.Poll(retryInterval, timeout, func() (bool, error) {
		tlpStress := &v1alpha1.TLPStress{}
		err := f.Client.Get(goctx.TODO(), types.NamespacedName{Namespace: namespace, Name: name}, tlpStress)
		if err != nil {
			if apierrors.IsNotFound(err) {
				t.Logf("Waiting for availability of TLPStress %s\n", name)
				return false, nil
			}
			return false, err
		}
		if tlpStress.Status.JobStatus == nil {
			t.Log("TLPStress.Status.JobStatus is nil")
			return false, nil
		}
		if tlpStress.Status.JobStatus.Active == 0 {
			t.Logf("Waiting for TLPStress %s to start\n", name)
			return false, nil
		}
		return true, nil
	})
}

// Waits for the TLPStress instance specified by namespace/name to finish. Specifically this
// function blocks until either TLPStress.Status.JobStatus.Succeeded > 0 or
// TLPStress.Status.JobStatus.Failed > 0. Note that this function currently assumes a single
// job instance is run.
func WaitForTLPStressToFinish(t *testing.T,
	f *framework.Framework,
	namespace string,
	name string,
	retryInterval time.Duration,
	timeout time.Duration,) error {

	return wait.Poll(retryInterval, timeout, func() (bool, error) {
		tlpStress := &v1alpha1.TLPStress{}
		err := f.Client.Get(goctx.TODO(), types.NamespacedName{Namespace: namespace, Name: name}, tlpStress)
		if err != nil {
			if apierrors.IsNotFound(err) {
				t.Logf("Waiting for availability of TLPStress %s\n", name)
				return false, nil
			}
			return false, err
		}
		if tlpStress.Status.JobStatus == nil {
			t.Log("TLPStress.Status.JobStatus is nil")
			return false, nil
		}

		if tlpStress.Status.JobStatus.Succeeded == 0 && tlpStress.Status.JobStatus.Failed == 0 {
			t.Logf("Waiting for TLPStress %s to finish, succeeded(%d), failed(%d)\n", name,
				tlpStress.Status.JobStatus.Succeeded, tlpStress.Status.JobStatus.Failed)
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
