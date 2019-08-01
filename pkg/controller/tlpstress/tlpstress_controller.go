package tlpstress

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/types"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"strconv"
	"strings"

	thelastpicklev1alpha1 "github.com/jsanda/tlp-stress-operator/pkg/apis/thelastpickle/v1alpha1"
	v1batch "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_tlpstress")

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
		if err = r.client.Status().Update(context.TODO(), tlpStress); err != nil {
			return reconcile.Result{}, err
		}
		return reconcile.Result{Requeue: true}, nil
	}

	// Check if the job already exists, if not create a new one
	job := &v1batch.Job{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: tlpStress.Name, Namespace: tlpStress.Namespace}, job)
	if err != nil && errors.IsNotFound(err) {
		// Define a new newJob
		newJob := r.jobForTLPStress(tlpStress, reqLogger)
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

func (r *ReconcileTLPStress) jobForTLPStress(tlpStress *thelastpicklev1alpha1.TLPStress, log logr.Logger) *v1batch.Job {
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
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					RestartPolicy: corev1.RestartPolicyOnFailure,
					Containers: []corev1.Container{
						{
							Name: tlpStress.Name,
							Image: tlpStress.Spec.Image,
							ImagePullPolicy: tlpStress.Spec.ImagePullPolicy,
							Args: *buildCmdLineArgs(tlpStress, log),
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

func buildCmdLineArgs(tlpStress  *thelastpicklev1alpha1.TLPStress, log logr.Logger) *[]string {
	args := make([]string, 0)

	args = append(args, "run", tlpStress.Spec.Workload)

	if len(tlpStress.Spec.ConsistencyLevel) > 0 {
		args = append(args, "--cl")
		args = append(args, tlpStress.Spec.ConsistencyLevel)
	}

	if tlpStress.Spec.Partitions != nil {
		args = append(args, "-p")
		args = append(args, strconv.FormatInt(*tlpStress.Spec.Partitions, 10))
	}

	if len(tlpStress.Spec.Duration) > 0 {
		args = append(args, "-d")
		args = append(args, tlpStress.Spec.Duration)
	}

	if tlpStress.Spec.DropKeyspace != nil && *tlpStress.Spec.DropKeyspace{
		args = append(args, "--drop")
	}

	if tlpStress.Spec.Iterations != nil && *tlpStress.Spec.Iterations != 0 {
		args = append(args, "-n")
		args = append(args, strconv.FormatInt(*tlpStress.Spec.Iterations, 10))
	}

	if len(tlpStress.Spec.ReadRate) > 0 {
		args = append(args, "-r")
		args = append(args, tlpStress.Spec.ReadRate)
	}

	if tlpStress.Spec.Populate != nil && *tlpStress.Spec.Populate != 0 {
		args = append(args, "--populate")
		args = append(args, strconv.FormatInt(*tlpStress.Spec.Populate, 10))
	}

	if tlpStress.Spec.Concurrency != nil && *tlpStress.Spec.Concurrency != 100 {
		args = append(args, "-c")
		args = append(args, strconv.FormatInt(int64(*tlpStress.Spec.Concurrency), 10))
	}

	if len(tlpStress.Spec.PartitionGenerator) > 0 {
		args = append(args, "--pg")
		args = append(args, tlpStress.Spec.PartitionGenerator)
	}

	// TODO Need to make sure only one replication strategy is specified
	if tlpStress.Spec.Replication.SimpleStrategy != nil {
		replicationFactor := strconv.FormatInt(int64(*tlpStress.Spec.Replication.SimpleStrategy), 10)
		replication := fmt.Sprintf(`{'class': 'SimpleStrategy', 'replication_factor': %s}`, replicationFactor)
		args = append(args, "--replication")
		args = append(args, replication)
	} else if tlpStress.Spec.Replication.NetworkTopologyStrategy != nil {
		var sb strings.Builder
		dcs := make([]string, 0)
		for k, v := range *tlpStress.Spec.Replication.NetworkTopologyStrategy {
			sb.WriteString("'")
			sb.WriteString(k)
			sb.WriteString("': ")
			sb.WriteString(strconv.FormatInt(int64(v), 10))
			dcs = append(dcs, sb.String())
			sb.Reset()
		}
		replication := fmt.Sprintf("{'class': 'NetworkTopologyStrategy', %s}", strings.Join(dcs, ", "))
		args = append(args, "--replication")
		args = append(args, replication)
	}

	// TODO add validation check
	args = append(args, "--host")
	args = append(args, tlpStress.Spec.CassandraService)

	return &args
}

func checkDefaults(tlpStress *thelastpicklev1alpha1.TLPStress) bool {
	updated := false

	if len(tlpStress.Spec.Image) == 0 {
		tlpStress.Spec.Image = "jsanda/tlp-stress:demo"
		updated = true
	}

	if len(tlpStress.Spec.ImagePullPolicy) == 0 {
		tlpStress.Spec.ImagePullPolicy = corev1.PullAlways
		updated = true
	}

	if len(tlpStress.Spec.Workload) == 0 {
		tlpStress.Spec.Workload = "KeyValue"
		updated = true
	}

	return updated
}

func labelsForTLPStress(name string) map[string]string {
	return map[string]string{"app": "TLPStress", "tlp-stress": name}
}
