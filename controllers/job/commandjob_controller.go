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
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"

	zap "go.uber.org/zap"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	jobv1alpha1 "github.com/fennec-project/snoopy-operator/apis/job/v1alpha1"
)

// CommandJobReconciler reconciles a CommandJob object
type CommandJobReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Cmd    *jobv1alpha1.CommandJob
}

//+kubebuilder:rbac:groups=job.fennecproject.io,resources=commandjobs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=job.fennecproject.io,resources=commandjobs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=job.fennecproject.io,resources=commandjobs/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the CommandJob object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.9.2/pkg/reconcile
func (r *CommandJobReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// Log User Info message about new jobs
	logger.Info("Checking for new command Jobs")

	// get CommandJobs
	r.Cmd = &jobv1alpha1.CommandJob{}
	err := r.Client.Get(ctx, req.NamespacedName, r.Cmd)
	if err != nil {
		return ctrl.Result{}, err
	}
	// Log Debug message with full new Command Job values

	// Check for deletion timestamp and finalizers
	finalizer := "commandjob.job.fennecproject.io"
	if r.Cmd.ObjectMeta.DeletionTimestamp.IsZero() {

		// CommandJob is not being deleted, so if it does not have our finalizer,
		// then lets add the finalizer and update the object. This is equivalent
		// registering our finalizer.

		if !containsString(r.Cmd.GetFinalizers(), finalizer) {

			logger.Info("New CommandJob found setting finalizers")

			r.Cmd.SetFinalizers(append(r.Cmd.GetFinalizers(), finalizer))
			if err := r.Update(context.Background(), r.Cmd); err != nil {
				return ctrl.Result{Requeue: true}, err
			}
			// Running reconciliation tasks
			// Target pod list by label and namespace
			podlist, err := r.GetRunningPodsByLabel(ctx, r.Cmd.Spec.LabelSelector, r.Cmd.Spec.TargetNamespace)
			if err != nil {
				fmt.Println(err.Error())
				return ctrl.Result{Requeue: true}, err
			}

			// CronJob creation by target pod
			for _, pod := range podlist.Items {

				// Build the command with arguments for podtracer

				podtracerArgs := []string{}
				podtracerArgs = append(podtracerArgs, "run")
				podtracerArgs = append(podtracerArgs, r.Cmd.Spec.Command)
				podtracerArgs = append(podtracerArgs, "-a")
				podtracerArgs = append(podtracerArgs, r.Cmd.Spec.Args)
				podtracerArgs = append(podtracerArgs, "--pod")
				podtracerArgs = append(podtracerArgs, pod.ObjectMeta.Name)
				podtracerArgs = append(podtracerArgs, "-n")
				podtracerArgs = append(podtracerArgs, pod.ObjectMeta.Namespace)

				// Temporarily listen to messages from podtracer on the operator pod
				// with nc and write those to file. Get the operator pod's IP and serve on port 5555 for now.

				// Generate the Cronjob object
				CronJob, err := r.CronJob(podtracerArgs, pod.ObjectMeta.Name, pod.Spec.NodeName, r.Cmd.Spec.Schedule)
				if err != nil {
					fmt.Println(err.Error())

					return ctrl.Result{}, err
				}

				// Create the Job in k8s api
				err = r.Client.Create(context.Background(), CronJob)
				if err != nil {
					fmt.Println(err.Error())
					return ctrl.Result{}, err
				}
				r.Cmd.Status.CronJobList = append(r.Cmd.Status.CronJobList, CronJob.ObjectMeta.Name)
				err = r.Client.Status().Update(context.Background(), r.Cmd)
				if err != nil {
					logger.Info(err.Error())
					return ctrl.Result{Requeue: true}, nil
				}
			}

		} else {
			// CommandJob is being deleted
			if containsString(r.Cmd.GetFinalizers(), finalizer) {

				// Find the list of snoopy jobs created on the status field
				// Delete all of them and set

				// remove our finalizer from the list and update it.
				r.Cmd.SetFinalizers(removeString(r.Cmd.GetFinalizers(), finalizer))
				if err := r.Update(context.Background(), r.Cmd); err != nil {
					return ctrl.Result{Requeue: true}, err
				}
			}
		}
	}

	// Log User Info message when job is being deleted or registering finalizer

	// List Target Pods matching labels on selected namespace

	return ctrl.Result{Requeue: false}, nil
}

func (r *CommandJobReconciler) GetRunningPodsByLabel(ctx context.Context, label map[string]string, namespace string) (*corev1.PodList, error) {

	podlist := &corev1.PodList{}
	listOpts := []client.ListOption{
		client.MatchingLabels(label),
		client.InNamespace(namespace),
		// client.MatchingFields{"phase": "Running"}, // TODO: TSHOOT contantly returning status.phase doesn't exist...
	}

	err := r.Client.List(ctx, podlist, listOpts...)
	if err != nil {
		fmt.Printf("GetRunningPodsByLabel, Error listing pods for tcpdump: %s ", err.Error())
		return nil, err
	}

	if len(podlist.Items) <= 0 {
		return nil, fmt.Errorf("no running pod corresponds to label %v and namespace %v ", label, namespace)
	}

	return podlist, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CommandJobReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&jobv1alpha1.CommandJob{}).
		Complete(r)
}

// Helper functions to check and remove string from a slice of strings.
func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func removeString(slice []string, s string) (result []string) {
	for _, item := range slice {
		if item == s {
			continue
		}
		result = append(result, item)
	}
	return
}
