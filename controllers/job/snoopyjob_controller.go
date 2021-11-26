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
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
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
<<<<<<< HEAD

	snoopyJob := &jobv1alpha1.SnoopyJob{}

	err := r.Client.Get(ctx, req.NamespacedName, snoopyJob)
=======

	// get SnoopyJobs
	job := &jobv1alpha1.SnoopyJob{}
	err := r.Client.Get(ctx, req.NamespacedName, job)
>>>>>>> a3ed996fe7c5677a2ca1789929664786f7ba8bdc
	if err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}

		return ctrl.Result{}, err
	}
<<<<<<< HEAD
=======

	cronJobs, err := r.BuildCronJobForPods(job)
	if err != nil {
		return ctrl.Result{Requeue: true}, err
	}

	if err = r.ReconcileCronJobs(cronJobs); err != nil {
		return ctrl.Result{Requeue: true}, err
	}

	// ****** BuildJobForPods(podlist *corev1.PodList, podtracerOpts []string) *[]batchv1.Job

	return ctrl.Result{Requeue: false}, nil
}

func (r *SnoopyJobReconciler) ReconcileCronJobs(cronJobs *batchv1.CronJobList) error {

	for _, cronJob := range cronJobs.Items {
		// Create the Job in k8s api
		err := r.Client.Create(context.Background(), &cronJob)
		if err != nil {
			fmt.Println(err.Error())
			return err
		}

		// Updating Status field after creating
		r.Cmd.Status.CronJobList = append(r.Cmd.Status.CronJobList, cronJob.ObjectMeta.Name)
		err = r.Client.Status().Update(context.Background(), r.Cmd)
		if err != nil {
			return nil
		}
	}

	return nil
}

func (r *SnoopyJobReconciler) BuildCronJobForPods(job *jobv1alpha1.SnoopyJob) (*batchv1.CronJobList, error) {

	// Running reconciliation tasks
	// Target pod list by label and namespace
	podlist, err := r.GetRunningPodsByLabel(context.TODO(), job.Spec.LabelSelector, job.Spec.TargetNamespace)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	cronJobs := &batchv1.CronJobList{}
	// CronJob creation by target pod
	for _, pod := range podlist.Items {

		// Build the command with arguments for podtracer

		podtracerOpts := r.buildPodtracerOptions()

		podtracerOpts = append(podtracerOpts, "--pod")
		podtracerOpts = append(podtracerOpts, pod.ObjectMeta.Name)
		podtracerOpts = append(podtracerOpts, "-n")
		podtracerOpts = append(podtracerOpts, pod.ObjectMeta.Namespace)

		// Temporarily listen to messages from podtracer on the operator pod
		// with nc and write those to file. Get the operator pod's IP and serve on port 5555 for now.

		// Generate the Cronjob object
		cronJob, err := r.CronJob(podtracerOpts, pod.ObjectMeta.Name, pod.Spec.NodeName, r.Cmd.Spec.Schedule)
		if err != nil {
			return nil, err
		}
		ctrl.SetControllerReference(job, cronJob, r.Scheme)
		cronJobs.Items = append(cronJobs.Items, *cronJob)
	}
	return cronJobs, nil
}

func (r *SnoopyJobReconciler) buildPodtracerOptions() []string {

	podtracerOpts := []string{}
	podtracerOpts = append(podtracerOpts, "run")
	podtracerOpts = append(podtracerOpts, r.Cmd.Spec.Command)
	podtracerOpts = append(podtracerOpts, "-a")
	podtracerOpts = append(podtracerOpts, r.Cmd.Spec.Args)

	if r.Cmd.Spec.Timer != "" {
		podtracerOpts = append(podtracerOpts, "-t")
		podtracerOpts = append(podtracerOpts, r.Cmd.Spec.Timer)
	}

	return podtracerOpts
}

func (r *SnoopyJobReconciler) GetRunningPodsByLabel(ctx context.Context, label map[string]string, namespace string) (*corev1.PodList, error) {

	podlist := &corev1.PodList{}
	listOpts := []client.ListOption{
		client.MatchingLabels(label),
		client.InNamespace(namespace),
		// client.MatchingFields{"phase": "Running"}, // TODO: TSHOOT contantly returning status.phase doesn't exist...
	}
>>>>>>> a3ed996fe7c5677a2ca1789929664786f7ba8bdc

	cronJobs, err := r.buildCronJobForPods(snoopyJob)
	if err != nil {
		return ctrl.Result{Requeue: true}, err
	}

	if err = r.reconcileCronJobs(snoopyJob, cronJobs); err != nil {
		return ctrl.Result{Requeue: true}, err
	}

	// ****** BuildJobForPods(podlist *corev1.PodList, podtracerOpts []string) *[]batchv1.Job

	return ctrl.Result{Requeue: false}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *SnoopyJobReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&jobv1alpha1.SnoopyJob{}).
		Owns(&batchv1.CronJob{}).
		Owns(&batchv1.Job{}).
<<<<<<< HEAD
		WithOptions(controller.Options{MaxConcurrentReconciles: 1}).
=======
>>>>>>> a3ed996fe7c5677a2ca1789929664786f7ba8bdc
		Complete(r)
}
