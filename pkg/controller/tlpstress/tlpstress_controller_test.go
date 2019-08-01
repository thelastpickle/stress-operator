package tlpstress

import (
	"github.com/jsanda/tlp-stress-operator/pkg/apis/thelastpickle/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"testing"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestTLPStressControllerJobCreate(t *testing.T) {
	var (
		name = "tlpstress-controller"
		namespace = "tlpstress"
	)

	tlpStress := &v1alpha1.TLPStress{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Namespace: namespace,
		},
		Spec: v1alpha1.TLPStressSpec{
			CassandraService: "cassandra-service",
		},
	}

	objs := []runtime.Object{ tlpStress }

	s := scheme.Scheme
	s.AddKnownTypes(v1alpha1.SchemeGroupVersion, tlpStress)

	cl := fake.NewFakeClient(objs...)

	r := &ReconcileTLPStress{client: cl, scheme: s}

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name: name,
			Namespace: namespace,
		},
	}

	res, err := r.Reconcile(req)
	if err != nil {
		t.Fatalf("reconcile: (%v)", err)
	}
	// Check the result of reconciliation to make sure it has the desired state.
	if !res.Requeue {
		t.Error("reconcile did not requeue request as expected")
	}
}