/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package job

import (
	"context"

	jobv1alpha1 "github.com/fennec-project/snoopy-operator/apis/job/v1alpha1"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
)

// SnoopyJobReconciler reconciles a SnoopyJob object
type SnoopyJobReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=job.fennecproject.io,resources=snoopyjobs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=job.fennecproject.io,resources=snoopyjobs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=job.fennecproject.io,resources=snoopyjobs/finalizers,verbs=update

func (r *SnoopyJobReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	snoopyJob := &jobv1alpha1.SnoopyJob{}

	err := r.Client.Get(ctx, req.NamespacedName, snoopyJob)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, err
	}

	if snoopyJob.Spec.Schedule != "" {

		cronJobs, err := r.buildCronJobForPods(snoopyJob)
		if err != nil {
			return ctrl.Result{Requeue: true}, err
		}

		if err = r.reconcileCronJobs(snoopyJob, cronJobs); err != nil {
			return ctrl.Result{Requeue: true}, err
		}

	} else {

		jobs, err := r.buildJobForPods(snoopyJob)
		if err != nil {
			return ctrl.Result{Requeue: true}, err
		}

		if err = r.reconcileJobs(snoopyJob, jobs); err != nil {
			return ctrl.Result{Requeue: true}, err
		}
	}

	return ctrl.Result{Requeue: false}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *SnoopyJobReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&jobv1alpha1.SnoopyJob{}).
		Owns(&batchv1.CronJob{}).
		Owns(&batchv1.Job{}).
		WithOptions(controller.Options{MaxConcurrentReconciles: 1}).
		Complete(r)
}
