package tlpstress

import (
	"context"
	"fmt"
	casskop "github.com/Orange-OpenSource/cassandra-k8s-operator/pkg/apis/db/v1alpha1"
	"github.com/go-logr/logr"
	thelastpicklev1alpha1 "github.com/jsanda/tlp-stress-operator/pkg/apis/thelastpickle/v1alpha1"
	"github.com/jsanda/tlp-stress-operator/pkg/tlpstress"
	v1batch "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
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
	DefaultWorkload        = thelastpicklev1alpha1.KeyValueWorkload
)

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new TLPStress Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
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
	err = c.Watch(&source.Kind{Type: &thelastpicklev1alpha1.TLPStress{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner TLPStress
	err = c.Watch(&source.Kind{Type: &v1batch.Job{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &thelastpicklev1alpha1.TLPStress{},
	})
	if err != nil {
		return err
	}

	mgr.GetScheme().AddKnownTypes(schema.GroupVersion{Group: "db.orange.com", Version: "v1alpha1"},
		&casskop.CassandraCluster{},
		&casskop.CassandraClusterList{},
		&metav1.ListOptions{},
	)

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
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileTLPStress) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling TLPStress")

	// Fetch the TLPStress tlpStress
	instance := &thelastpicklev1alpha1.TLPStress{}
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

	// If a CassandraClusterTemplate is defined then make sure that:
	//    1) A CassandraCluster matching the template exists
	//    2) Create the CassandraCluster if it does not exist
	//    3) CassandraCluster.status.phase == Running
	if tlpStress.Spec.CassandraConfig.CassandraClusterTemplate != nil {
		template := tlpStress.Spec.CassandraConfig.CassandraClusterTemplate

		if len(template.Namespace) == 0 {
			template.Namespace = request.Namespace
		}

		cc := &casskop.CassandraCluster{}
		if err = r.client.Get(context.TODO(), types.NamespacedName{Name: template.Name, Namespace: template.Namespace}, cc); err != nil {
			if !errors.IsNotFound(err) {
				return reconcile.Result{RequeueAfter: 5 * time.Second}, err
			}

			cc.ObjectMeta = template.ObjectMeta
			cc.TypeMeta = template.TypeMeta
			cc.Spec = template.Spec

			reqLogger.Info("Creating a new CassandraCluster.", "CassandraCluster.Namespace",
				cc.Namespace, "CassandraCluster.Name", cc.Name)
			if err = r.client.Create(context.TODO(), cc); err != nil {
				return reconcile.Result{RequeueAfter: 5 * time.Second}, err
			} else {
				return reconcile.Result{Requeue: true}, nil
			}
		} else {
			if cc.Status.Phase != "Running" {
				reqLogger.Info("Waiting for CassandraCluster to be ready.", "CassandraCluster.Name",
					cc.Name, "CassandraCluster.Status.Phase", cc.Status.Phase)
				return reconcile.Result{RequeueAfter: 30 * time.Second}, nil
			}
		}
	}

	// Check if the metrics service already exists, if not create a new one
	metricsService := &corev1.Service{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Namespace: tlpStress.Namespace, Name: getMetricsServiceName(tlpStress)}, metricsService)
	if err != nil && errors.IsNotFound(err) {
		// Define a new service
		metricsService = r.createMetricsService(tlpStress, request.Namespace)
		reqLogger.Info("Creating metrics service.", "MetricsService.Namespace", metricsService.Namespace,
			"MetricsService.Name", metricsService.Name)
		err = r.client.Create(context.TODO(), metricsService)
		if err != nil {
			reqLogger.Error(err, "Failed to create metrics service.", "MetricsService.Namespace",
				metricsService.Namespace, "MetricsService.Name", metricsService.Name)
			return reconcile.Result{}, err
		}
		return reconcile.Result{Requeue: true}, nil
	} else if err != nil {
		reqLogger.Error(err,"Failed to get MetricsService", "MetricsService.Namespace",
			tlpStress.Namespace, "MetricsService.Name", getMetricsServiceName(tlpStress))
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

func (r *ReconcileTLPStress) createMetricsService(tlpStress *thelastpicklev1alpha1.TLPStress, namespace string) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: getMetricsServiceName(tlpStress),
			Namespace: namespace,
			Labels: labelsForTLPStress(tlpStress.Name),
			OwnerReferences: []metav1.OwnerReference{
				tlpStress.CreateOwnerReference(),
			},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Port: 9500,
					Name: "metrics",
					Protocol: corev1.ProtocolTCP,
					TargetPort: intstr.IntOrString{
						Type: intstr.Int,
						IntVal: 9500,
					},
				},
			},
			Selector: labelsForTLPStress(tlpStress.Name),
		},
	}
}

func getMetricsServiceName(tlpStress *thelastpicklev1alpha1.TLPStress) string {
	return fmt.Sprintf("%s-metrics", tlpStress.Name)
}

func (r *ReconcileTLPStress) jobForTLPStress(tlpStress *thelastpicklev1alpha1.TLPStress, namespace string,
	log logr.Logger) *v1batch.Job {

	ls := labelsForTLPStress(tlpStress.Name)

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

func buildCmdLineArgs(instance *thelastpicklev1alpha1.TLPStress, namespace string, log logr.Logger) *[]string {
	cmdLineArgs := tlpstress.CreateCommandLineArgs(&instance.Spec.StressConfig, &instance.Spec.CassandraConfig, namespace)
	log.Info("Creating tlp-stress arguments", "commandLineArgs", cmdLineArgs)
	return cmdLineArgs.GetArgs()
}

func checkDefaults(tlpStress *thelastpicklev1alpha1.TLPStress) bool {
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

func labelsForTLPStress(name string) map[string]string {
	return map[string]string{
		"app": "tlpstress",
		"tlpstress": name,
		"prometheus": "enabled",
	}
}
