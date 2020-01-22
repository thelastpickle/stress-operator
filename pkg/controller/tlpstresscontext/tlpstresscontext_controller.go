package tlpstresscontext

import (
	"context"
	thelastpicklev1alpha1 "github.com/jsanda/tlp-stress-operator/pkg/apis/thelastpickle/v1alpha1"
	"github.com/jsanda/tlp-stress-operator/pkg/monitoring"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_tlpstresscontext")

const contextName = "tlpstress"

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new TLPStressContext Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileTLPStressContext{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("tlpstresscontext-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource TLPStressContext
	err = c.Watch(&source.Kind{Type: &thelastpicklev1alpha1.TLPStressContext{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner TLPStressContext
	//err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
	//	IsController: true,
	//	OwnerType:    &thelastpicklev1alpha1.TLPStressContext{},
	//})
	//if err != nil {
	//	return err
	//}

	return nil
}

// blank assignment to verify that ReconcileTLPStressContext implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileTLPStressContext{}

// ReconcileTLPStressContext reconciles a TLPStressContext object
type ReconcileTLPStressContext struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a TLPStressContext object and makes changes based on the state read
// and what is in the TLPStressContext.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileTLPStressContext) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling TLPStressContext")

	if request.Name != contextName {
		reqLogger.Info("Ignoring request", "Request.Namespace", request.Namespace,
			"Request.Name", request.Name)
		return reconcile.Result{}, nil
	}

	// Fetch the TLPStressContext instance
	instance := &thelastpicklev1alpha1.TLPStressContext{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	stressContext := instance.DeepCopy()

	if stressContext.Spec.InstallPrometheus {
		if kindExists, err := monitoring.PrometheusKindExists(); kindExists {
			_, err := monitoring.GetPrometheus(request.Namespace, r.client)
			if err != nil && errors.IsNotFound(err) {
				return monitoring.CreatePrometheus(request.Namespace, r.client, reqLogger)
			} else if err != nil {
				reqLogger.Error(err, "Failed to get Prometheus")
				return reconcile.Result{}, err
			}
		} else if err != nil {
			reqLogger.Error(err, "Failed to check for Prometheus CRD")
			return reconcile.Result{}, err
		}

		_, err := monitoring.GetPrometheusService(request.Namespace, r.client)
		if err != nil && errors.IsNotFound(err) {
			return monitoring.CreatePrometheusService(request.Namespace, r.client, reqLogger)
		} else if err != nil {
			reqLogger.Error(err, "Failed to get Prometheus service")
			return reconcile.Result{}, err
		}

		if kindExists, err := monitoring.ServiceMonitorKindExists(); kindExists {
			_, err := monitoring.GetServiceMonitor(request.Namespace, r.client)
			if err != nil && errors.IsNotFound(err) {
				return monitoring.CreateServiceMonitor(request.Namespace, r.client, reqLogger)
			} else if err != nil {
				reqLogger.Error(err, "Failed to get ServiceMonitor")
				return reconcile.Result{}, err
			}
		} else if err != nil {
			reqLogger.Error(err, "Failed to check for ServiceMonitor CRD")
			return reconcile.Result{}, err
		}
	}

	if stressContext.Spec.InstallGrafana {
		if kindExists, err := monitoring.GrafanaKindExists(); kindExists {
			_, err := monitoring.GetGrafana(request.Namespace, r.client)
			if err != nil && errors.IsNotFound(err) {
				return monitoring.CreateGrafana(request.Namespace, r.client, reqLogger)
			} else if err != nil {
				reqLogger.Error(err, "Failed to get Grafana")
				return reconcile.Result{}, err
			}
		} else if err != nil {
			reqLogger.Error(err, "Failed to check for Grafana CRD")
			return reconcile.Result{}, err
		}

		_, err = monitoring.GetDataSource(request.Namespace, r.client)
		if err != nil && errors.IsNotFound(err) {
			return monitoring.CreateDataSource(request.Namespace, r.client, reqLogger)
		} else if err != nil {
			reqLogger.Error(err, "Failed to get DataSource")
			return reconcile.Result{}, err
		}
	}

	return reconcile.Result{}, nil
}
