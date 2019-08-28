package e2eutil

import (
	goctx "context"
	casskop "github.com/Orange-OpenSource/cassandra-k8s-operator/pkg/apis/db/v1alpha1"
	framework "github.com/operator-framework/operator-sdk/pkg/test"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"testing"
	"time"
	"github.com/operator-framework/operator-sdk/pkg/test/e2eutil"
)

func WaitForCassKopCluster(
	t *testing.T,
	f *framework.Framework,
	namespace string,
	name string,
	retryInterval time.Duration,
	timeout time.Duration,) error {

	return wait.Poll(retryInterval, timeout, func() (done bool, err error) {
		cc := &casskop.CassandraCluster{}
		getErr := f.Client.Get(goctx.TODO(), types.NamespacedName{Name: name, Namespace: namespace}, cc)
		if getErr != nil {
			if apierrors.IsNotFound(getErr) {
				t.Logf("Waiting for availability of CassandraCluster %s\n", name)
				return false, nil
			}
			return false, getErr
		}
		if (cc.Status.Phase != "Running") {
			t.Logf("Waiting for CassandraCassandra %s (%s)\n", name, cc.Status.Phase)
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
