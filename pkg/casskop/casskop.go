package casskop

import (
	"context"
	casskopapi "github.com/Orange-OpenSource/cassandra-k8s-operator/pkg/apis"
	casskop "github.com/Orange-OpenSource/cassandra-k8s-operator/pkg/apis/db/v1alpha1"
	"github.com/go-logr/logr"
	api "github.com/thelastpickle/stress-operator/pkg/apis/thelastpickle/v1alpha1"
	"github.com/thelastpickle/stress-operator/pkg/k8s"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"time"
)

const CassandraClusterKind = "CassandraCluster"

var discoveryClient k8s.DiscoveryClient

func Init(dc k8s.DiscoveryClient) {
	discoveryClient = dc
}

func AddToScheme(scheme *runtime.Scheme) error {
	return casskopapi.AddToScheme(scheme)
	//return casskop.SchemeBuilder.AddToScheme(scheme)
	//return nil
}

func CassandraClusterKindExists() (bool, error) {
	return discoveryClient.KindExists(casskop.SchemeGroupVersion.String(), CassandraClusterKind)
}

func GetCassandraCluster(template *api.CassandraClusterTemplate, client client.Client) (*casskop.CassandraCluster, error) {
	cc := &casskop.CassandraCluster{}
	err := client.Get(context.TODO(), types.NamespacedName{Name: template.Name, Namespace: template.Namespace}, cc)
	return cc, err
}

func CreateCassandraCluster(template *api.CassandraClusterTemplate, client client.Client, log logr.Logger) (reconcile.Result, error) {
	cc := &casskop.CassandraCluster{}
	cc.ObjectMeta = template.ObjectMeta
	cc.TypeMeta = template.TypeMeta
	cc.Spec = template.Spec

	log.Info("Creating a new CassandraCluster.", "CassandraCluster.Namespace", cc.Namespace, "CassandraCluster.Name", cc.Name)
	if err := client.Create(context.TODO(), cc); err != nil {
		return reconcile.Result{RequeueAfter: 5 * time.Second}, err
	} else {
		return reconcile.Result{Requeue: true}, nil
	}
}

func IsCassandraClusterReady(cc *casskop.CassandraCluster) bool {
	return cc.Status.Phase == "Running"
}
