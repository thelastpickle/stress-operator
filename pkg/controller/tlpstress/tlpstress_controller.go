package tlpstress

import (
	"context"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

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

const (
	PhaseInitialize int = iota
	PhaseJobCreate
	PhaseJobRunning
	PhaseJobDone
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
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	tlpStress := instance.DeepCopy()

	switch tlpStress.Status.Phase {
	case PhaseInitialize:
		return r.checkDefaults(tlpStress)
	case PhaseJobCreate:
		return r.createJobIfNecessary(tlpStress, reqLogger)
	}

	return reconcile.Result{Requeue: true}, nil
}

func (r *ReconcileTLPStress) checkDefaults(tlpStress *thelastpicklev1alpha1.TLPStress) (reconcile.Result, error) {
	checkDefaults(tlpStress)
	tlpStress.Status.Phase = PhaseJobCreate
	err := r.client.Update(context.TODO(), tlpStress)
	return reconcile.Result{Requeue: true}, err
}

func (r *ReconcileTLPStress) createJobIfNecessary(tlpStress *thelastpicklev1alpha1.TLPStress, logger logr.Logger) (reconcile.Result, error) {
	// Check if the job already exists, if not create a new one
	found := &v1batch.Job{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: tlpStress.Name, Namespace: tlpStress.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		// Define a new job
		job := r.jobForTLPStress(tlpStress)
		logger.Info("Creating a new Job.", "Job.Namespace", job.Namespace, "Job.Name", job.Name)
		err = r.client.Create(context.TODO(), job)
		if err != nil {
			logger.Error(err,"Failed to create new Job.", "Job.Namespace", job.Namespace, "Job.Name", job.Name)
			return reconcile.Result{}, err
		}
		// Job created successfully - return and requeue
		tlpStress.Status.Phase = PhaseJobRunning
		return reconcile.Result{Requeue: true}, nil
	} else {
		logger.Error(err,"Failed to get Job.")
		return reconcile.Result{}, err
	}
}

func (r *ReconcileTLPStress) jobForTLPStress(tlpStress *thelastpicklev1alpha1.TLPStress) *v1batch.Job {
	ls := labelsForTLPStress(tlpStress.Name)

	job := &v1batch.Job{
		TypeMeta: metav1.TypeMeta{
			Kind: "Job",
			APIVersion: "v1",
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
							Args: []string {"run", tlpStress.Spec.Workload},
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

func (r *ReconcileTLPStress) updatePhase(cr *thelastpicklev1alpha1.TLPStress, phase int) (reconcile.Result, error) {
	cr.Status.Phase = phase
	err := r.client.Update(context.TODO(), cr)
	return reconcile.Result{}, err
}
