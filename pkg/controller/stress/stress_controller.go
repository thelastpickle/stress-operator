package stress

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	api "github.com/jsanda/stress-operator/pkg/apis/thelastpickle/v1alpha1"
	casskoputil "github.com/jsanda/stress-operator/pkg/casskop"
	"github.com/jsanda/stress-operator/pkg/monitoring"
	"github.com/jsanda/stress-operator/pkg/tlpstress"
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
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"time"
)

var log = logf.Log.WithName("controller_stress")

const (
	DefaultImage           = "thelastpickle/tlp-stress:4.0.0"
	DefaultImagePullPolicy = corev1.PullIfNotPresent
	DefaultWorkload        = api.KeyValueWorkload
)

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Stress Controller and adds it to the Manager. The Manager will set fields on the Controller
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
	return &ReconcileStress{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("stress-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Stress
	err = c.Watch(&source.Kind{Type: &api.Stress{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner Stress
	err = c.Watch(&source.Kind{Type: &v1batch.Job{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &api.Stress{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileStress implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileStress{}

// ReconcileStress reconciles a Stress object
type ReconcileStress struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Stress object and makes changes based on the state read
// and what is in the Stress.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.ReconcileRequeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileStress) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Stress")

	// Fetch the Stress instance
	instance := &api.Stress{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not job, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		reqLogger.Error(err, "Failed to get Stress object")
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	stress := instance.DeepCopy()

	if checkDefaults(stress) {
		if err = r.client.Update(context.TODO(), stress); err != nil {
			return reconcile.Result{}, err
		}
		return reconcile.Result{Requeue: true}, nil
	}

	// If the Stress specifies a CassandraCluster reference or a template, then we need to check that the CRD exists
	// on the master. If the CRD does not exist, then we just requeue with an error.
	if stress.Spec.CassandraConfig.CassandraCluster != nil || stress.Spec.CassandraConfig.CassandraClusterTemplate != nil {
		if kindExists, err := casskoputil.CassandraClusterKindExists(); !kindExists {
			reqLogger.Info("Cannot create Stress instance. The CassandraCluster kind does not exist.",
				"Stress.Name", stress.Name, "Stress.Namespace", stress.Namespace)
			return reconcile.Result{}, fmt.Errorf("cannot create Stress instance %s.%s: CassandraCluster kind does not exist",
				stress.Namespace, stress.Name)
		} else if err != nil {
			reqLogger.Error(err,"Check for CassandraCluster kind failed")
			return reconcile.Result{}, err
		} else {
			// If a CassandraClusterTemplate is defined then make sure that:
			//    1) A CassandraCluster matching the template exists
			//    2) Create the CassandraCluster if it does not exist
			//    3) CassandraCluster is ready
			if stress.Spec.CassandraConfig.CassandraClusterTemplate != nil {
				template := stress.Spec.CassandraConfig.CassandraClusterTemplate
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
	_, err = monitoring.GetMetricsService(stress, r.client)
	if err != nil && errors.IsNotFound(err) {
		return monitoring.CreateMetricsService(stress, r.client, reqLogger)
	} else if err != nil {
		reqLogger.Error(err,"Failed to get MetricsService", "MetricsService.Namespace",
			stress.Namespace, "MetricsService.Name", monitoring.GetMetricsServiceName(stress))
		return reconcile.Result{}, err
	}

	if kindExists, err := monitoring.GrafanaDashboardKindExists(); kindExists {
		_, err := monitoring.GetDashboard(stress, r.client)
		if err != nil && errors.IsNotFound(err) {
			// Create the dashboard
			return monitoring.CreateDashboard(stress, r.client, reqLogger)
		} else if err != nil {
			reqLogger.Error(err, "Failed to get dashboard", "GrafanaDashboard.Namespace",
				stress.Namespace, "GrafanaDashboard.Name", stress.Name)
			return reconcile.Result{}, err
		}
	} else if err != nil {
		reqLogger.Error(err, "Check for GrafanaDashboard kind failed")
		return reconcile.Result{}, err
	}

	// Check if the job already exists, if not create a new one
	job := &v1batch.Job{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: stress.Name, Namespace: stress.Namespace}, job)
	if err != nil && errors.IsNotFound(err) {
		// Define a new newJob
		newJob := r.jobForStress(stress, request.Namespace, reqLogger)
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
	if stress.Status.JobStatus == nil || !reflect.DeepEqual(stress.Status.JobStatus, jobStatus) {
		stress.Status.JobStatus = jobStatus
		if err = r.client.Status().Update(context.TODO(), stress); err != nil {
			reqLogger.Error(err, "Failed to update status")
			return reconcile.Result{}, err
		}
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileStress) jobForStress(stress *api.Stress, namespace string, log logr.Logger) *v1batch.Job {

	ls := tlpstress.LabelsForStress(stress.Name)

	job := &v1batch.Job{
		TypeMeta: metav1.TypeMeta{
			Kind: "Job",
			APIVersion: "batch/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      stress.Name,
			Namespace: stress.Namespace,
			Labels:    ls,
		},
		Spec: v1batch.JobSpec{
			BackoffLimit: stress.Spec.JobConfig.BackoffLimit,
			Parallelism:  stress.Spec.JobConfig.Parallelism,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ls,
				},
				Spec: corev1.PodSpec{
					RestartPolicy: corev1.RestartPolicyOnFailure,
					Containers: []corev1.Container{
						{
							Name:            stress.Name,
							Image:           stress.Spec.Image,
							ImagePullPolicy: stress.Spec.ImagePullPolicy,
							Args:            *buildCmdLineArgs(stress, namespace, log),
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
	// Set Stress as the owner and controller
	if err := controllerutil.SetControllerReference(stress, job, r.scheme); err != nil {
		// TODO We probably want to return and handle this error
		log.Error(err, "Failed to set owner for job")
	}

	return job
}

func buildCmdLineArgs(instance *api.Stress, namespace string, log logr.Logger) *[]string {
	cmdLineArgs := tlpstress.CreateCommandLineArgs(&instance.Spec.StressConfig, &instance.Spec.CassandraConfig, namespace)
	log.Info("Creating tlp-stress arguments", "commandLineArgs", cmdLineArgs)
	return cmdLineArgs.GetArgs()
}

func checkDefaults(stress *api.Stress) bool {
	updated := false

	if len(stress.Spec.Image) == 0 {
		stress.Spec.Image = DefaultImage
		updated = true
	}

	if len(stress.Spec.ImagePullPolicy) == 0 {
		stress.Spec.ImagePullPolicy = DefaultImagePullPolicy
		updated = true
	}

	if len(stress.Spec.StressConfig.Workload) == 0 {
		stress.Spec.StressConfig.Workload = DefaultWorkload
		updated = true
	}

	return updated
}
