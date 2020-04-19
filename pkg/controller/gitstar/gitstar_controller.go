package gitstar

import (
	"context"
	"fmt"

	"k8s.io/api/batch/v1"
	batchv1 "k8s.io/api/batch/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	appv1 "gitstar-operator/pkg/apis/app/v1"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const (
	ENVGitStarName      = "git_star_name"
	ENVGitStarNameSpace = "git_star_name_space"
)

var log = logf.Log.WithName("controller_gitstar")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new GitStar Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileGitStar{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("gitstar-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource GitStar
	err = c.Watch(&source.Kind{Type: &appv1.GitStar{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileGitStar implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileGitStar{}

// ReconcileGitStar reconciles a GitStar object
type ReconcileGitStar struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a GitStar object and makes changes based on the state read
// and what is in the GitStar.Spec
func (r *ReconcileGitStar) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling GitStar")

	// Fetch the GitStar instance
	instance := &appv1.GitStar{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// gitStar was delete

			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	cronJob := newCronJobForCR(instance)
	// Set GitStar instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, cronJob, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	found := &batchv1.CronJob{}
	err = r.client.Get(context.TODO(), request.NamespacedName, found)

	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new CronJob", "CronJob.Namespace", cronJob.Namespace, "CronJob.Name", cronJob.Name)
		err := r.client.Create(context.TODO(), cronJob)
		if err != nil {
			reqLogger.Error(err, "create CronJob of GetStar failed!")
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, nil
	} else if err != nil {
		// Error reading the object - requeue the request.
		reqLogger.Error(err, "get cronJob object failed!")
		return reconcile.Result{}, err
	}

	// Pod already exists - don't requeue
	reqLogger.Info("Skip reconcile: CronJob already exists", "CronJob.Namespace", found.Namespace, "CronJob.Name", found.Name)
	return reconcile.Result{}, nil
}

func generateCronJobName(cr *appv1.GitStar) string {
	return fmt.Sprintf("%s-gitstar", cr.Name)
}

func newCronJobForCR(cr *appv1.GitStar) *batchv1.CronJob {
	labels := map[string]string{
		"app": cr.Name,
	}
	historyLimit := int32(20)

	return &batchv1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      generateCronJobName(cr),
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: batchv1.CronJobSpec{
			Schedule: "12 * * * *", // every 1 hour
			JobTemplate: batchv1.JobTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{},
				Spec: v1.JobSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:  cr.Name + "-gitstarjob",
									Image: "",
									Env: []corev1.EnvVar{
										{
											Name:  ENVGitStarName,
											Value: cr.Spec.RepoName,
										},
										{
											Name:  ENVGitStarNameSpace,
											Value: cr.Namespace,
										},
									},
								},
							},
							RestartPolicy: corev1.RestartPolicyNever,
						},
					},
				},
			},
			SuccessfulJobsHistoryLimit: &historyLimit,
			FailedJobsHistoryLimit:     &historyLimit,
		},
	}
}
