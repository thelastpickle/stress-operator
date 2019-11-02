package tlpstress

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	api "github.com/jsanda/tlp-stress-operator/pkg/apis/thelastpickle/v1alpha1"
	casskoputil "github.com/jsanda/tlp-stress-operator/pkg/casskop"
	"github.com/jsanda/tlp-stress-operator/pkg/monitoring"
	"github.com/jsanda/tlp-stress-operator/pkg/tlpstress"
	v1batch "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"time"
)

var log = logf.Log.WithName("controller_tlpstress")

const (
	DefaultImage           = "thelastpickle/tlp-stress:3.0.0"
	DefaultImagePullPolicy = corev1.PullAlways
	DefaultWorkload        = api.KeyValueWorkload
)

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new TLPStress Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	//if dc, err := k8s.GetDiscoveryClient(); err != nil {
	//	return err
	//} else {
	//	monitoring.Init(dc)
	//	casskoputil.Init(dc)
	//}

	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileTLPStress{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("tlpstress-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource TLPStress
	err = c.Watch(&source.Kind{Type: &api.TLPStress{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner TLPStress
	err = c.Watch(&source.Kind{Type: &v1batch.Job{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &api.TLPStress{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileTLPStress implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileTLPStress{}

// ReconcileTLPStress reconciles a TLPStress object
type ReconcileTLPStress struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a TLPStress object and makes changes based on the state read
// and what is in the TLPStress.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.ReconcileRequeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileTLPStress) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling TLPStress")

	// Fetch the TLPStress tlpStress
	instance := &api.TLPStress{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not job, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		reqLogger.Error(err, "Failed to get TLPStress object")
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	tlpStress := instance.DeepCopy()

	if checkDefaults(tlpStress) {
		if err = r.client.Update(context.TODO(), tlpStress); err != nil {
			return reconcile.Result{}, err
		}
		return reconcile.Result{Requeue: true}, nil
	}

	// If the TLPStress specifies a CassandraCluster reference or a template, then we need to check that the CRD exists
	// on the master. If the CRD does not exist, then we just requeue with an error.
	if tlpStress.Spec.CassandraConfig.CassandraCluster != nil || tlpStress.Spec.CassandraConfig.CassandraClusterTemplate != nil {
		if kindExists, err := casskoputil.CassandraClusterKindExists(); !kindExists {
			reqLogger.Info("Cannot create TLPStress instance. The CassandraCluster kind does not exist.",
				"TLPStress.Name", tlpStress.Name, "TLPStress.Namespace", tlpStress.Namespace)
			return reconcile.Result{}, fmt.Errorf("cannot create TLPStress instance %s.%s: CassandraCluster kind does not exist",
				tlpStress.Namespace, tlpStress.Name)
		} else if err != nil {
			reqLogger.Error(err,"Check for CassandraCluster kind failed")
			return reconcile.Result{}, err
		} else {
			// If a CassandraClusterTemplate is defined then make sure that:
			//    1) A CassandraCluster matching the template exists
			//    2) Create the CassandraCluster if it does not exist
			//    3) CassandraCluster is ready
			if tlpStress.Spec.CassandraConfig.CassandraClusterTemplate != nil {
				template := tlpStress.Spec.CassandraConfig.CassandraClusterTemplate
				if len(template.Namespace) == 0 {
					template.Namespace = request.Namespace
				}
				cc, err := casskoputil.GetCassandraCluster(template, r.client)
				if err != nil {
					if !errors.IsNotFound(err) {
						return reconcile.Result{RequeueAfter: 5 * time.Second}, err
					}
					return casskoputil.CreateCassandraCluster(template, r.client, reqLogger)
				} else {
					if !casskoputil.IsCassandraClusterReady(cc) {
						reqLogger.Info("Waiting for CassandraCluster to be ready.", "CassandraCluster.Name",
							cc.Name, "CassandraCluster.Status.Phase", cc.Status.Phase)
						return reconcile.Result{RequeueAfter: 30 * time.Second}, nil
					}
				}
			}
		}
	}

	// Check if the metrics service already exists, if not create a new one
	metricsSvc, err := monitoring.GetMetricsService(tlpStress, r.client)
	if err != nil && errors.IsNotFound(err) {
		return monitoring.CreateMetricsService(tlpStress, r.client, reqLogger)
	} else if err != nil {
		reqLogger.Error(err,"Failed to get MetricsService", "MetricsService.Namespace",
			tlpStress.Namespace, "MetricsService.Name", monitoring.GetMetricsServiceName(tlpStress))
		return reconcile.Result{}, err
	}

	// The ServiceMonitor CRD is create by prometheus-operator which is an option dependency. We therefore
	// need to check that the CRD exists on the server before we try to create a ServiceMonitor.
	if crdDefined, err := monitoring.ServiceMonitorKindExists(); crdDefined {
		// Check if the service monitor already exists, if not create one
		serviceMonitor, err := monitoring.GetServiceMonitor(tlpStress, r.client)
		if err != nil && errors.IsNotFound(err) {
			// Define a new service monitor
			return monitoring.CreateServiceMonitor(metricsSvc, r.client, reqLogger)
		} else if err != nil {
			reqLogger.Error(err, "Failed to get ServiceMonitor", "ServiceMonitor.Namespace",
				serviceMonitor.Namespace, "ServiceMonitor.Name", serviceMonitor.Name)
			return reconcile.Result{}, err
		}
	} else if err != nil {
		reqLogger.Error(err, "Failed to check for ServiceMonitor CRD")
		return reconcile.Result{}, err
	}

	if kindExists, err := monitoring.GrafanaDashboardKindExists(); kindExists {
		_, err := monitoring.GetDashboard(tlpStress, r.client)
		if err != nil && errors.IsNotFound(err) {
			// Create the dashboard
			return monitoring.CreateDashboard(tlpStress, r.client, reqLogger)
		} else if err != nil {
			reqLogger.Error(err, "Failed to get dashboard", "GrafanaDashboard.Namespace",
				tlpStress.Namespace, "GrafanaDashboard.Name", tlpStress.Name)
			return reconcile.Result{}, err
		}
	} else if err != nil {
		reqLogger.Error(err, "Check for GrafanaDashboard kind failed")
		return reconcile.Result{}, err
	}

	// Check if the job already exists, if not create a new one
	job := &v1batch.Job{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: tlpStress.Name, Namespace: tlpStress.Namespace}, job)
	if err != nil && errors.IsNotFound(err) {
		// Define a new newJob
		newJob := r.jobForTLPStress(tlpStress, request.Namespace, reqLogger)
		reqLogger.Info("Creating a new Job.", "Job.Namespace", newJob.Namespace, "Job.Name", newJob.Name)
		err = r.client.Create(context.TODO(), newJob)
		if err != nil {
			reqLogger.Error(err,"Failed to create new Job.", "Job.Namespace", newJob.Namespace, "Job.Name", newJob.Name)
			return reconcile.Result{}, err
		}
		return reconcile.Result{Requeue: true}, nil
	} else if err != nil {
		reqLogger.Error(err,"Failed to get Job.")
		return reconcile.Result{}, err
	}

	// Check the status and update if it has changed
	jobStatus := job.Status.DeepCopy()
	if tlpStress.Status.JobStatus == nil || !reflect.DeepEqual(tlpStress.Status.JobStatus, jobStatus) {
		tlpStress.Status.JobStatus = jobStatus
		if err = r.client.Status().Update(context.TODO(), tlpStress); err != nil {
			reqLogger.Error(err, "Failed to update status")
			return reconcile.Result{}, err
		}
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileTLPStress) jobForTLPStress(tlpStress *api.TLPStress, namespace string,
	log logr.Logger) *v1batch.Job {

	ls := tlpstress.LabelsForTLPStress(tlpStress.Name)

	job := &v1batch.Job{
		TypeMeta: metav1.TypeMeta{
			Kind: "Job",
			APIVersion: "batch/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: tlpStress.Name,
			Namespace: tlpStress.Namespace,
			Labels: ls,
		},
		Spec: v1batch.JobSpec{
			BackoffLimit: tlpStress.Spec.JobConfig.BackoffLimit,
			Parallelism: tlpStress.Spec.JobConfig.Parallelism,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ls,
				},
				Spec: corev1.PodSpec{
					RestartPolicy: corev1.RestartPolicyOnFailure,
					Containers: []corev1.Container{
						{
							Name: tlpStress.Name,
							Image: tlpStress.Spec.Image,
							ImagePullPolicy: tlpStress.Spec.ImagePullPolicy,
							Args: *buildCmdLineArgs(tlpStress, namespace, log),
							Ports: []corev1.ContainerPort{
								{
									Name:          "metrics",
									ContainerPort: 9500,  // TODO make configurable
								},
							},
						},
					},

				},
			},
		},
	}
	// Set TLPStress as the owner and controller
	if err := controllerutil.SetControllerReference(tlpStress, job, r.scheme); err != nil {
		// TODO We probably want to return and handle this error
		log.Error(err, "Failed to set owner for job")
	}

	return job
}

func buildCmdLineArgs(instance *api.TLPStress, namespace string, log logr.Logger) *[]string {
	cmdLineArgs := tlpstress.CreateCommandLineArgs(&instance.Spec.StressConfig, &instance.Spec.CassandraConfig, namespace)
	log.Info("Creating tlp-stress arguments", "commandLineArgs", cmdLineArgs)
	return cmdLineArgs.GetArgs()
}

func checkDefaults(tlpStress *api.TLPStress) bool {
	updated := false

	if len(tlpStress.Spec.Image) == 0 {
		tlpStress.Spec.Image = DefaultImage
		updated = true
	}

	if len(tlpStress.Spec.ImagePullPolicy) == 0 {
		tlpStress.Spec.ImagePullPolicy = DefaultImagePullPolicy
		updated = true
	}

	if len(tlpStress.Spec.StressConfig.Workload) == 0 {
		tlpStress.Spec.StressConfig.Workload = DefaultWorkload
		updated = true
	}

	return updated
}
