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

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	jobv1alpha1 "github.com/fennec-project/snoopy-operator/apis/job/v1alpha1"
)

// CommandJobReconciler reconciles a CommandJob object
type CommandJobReconciler struct {
	client.Client
	Scheme *runtime.Scheme
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
	_ = log.FromContext(ctx)

	// get CommandJobs

	cmd := &jobv1alpha1.CommandJob{}
	err := r.Client.Get(ctx, req.NamespacedName, cmd)
	if err != nil {
		return ctrl.Result{}, err
	}

	// Log User Info message about new jobs

	// Log Debug message with full new Command Job values

	// Check for deletion timestamp and finalizers

	// Log User Info message when job is being deleted or registering finalizer

	// List Target Pods matching labels on selected namespace

	podlist, err := r.GetRunningPodsByLabel(ctx, cmd.Spec.LabelSelector, cmd.Spec.TargetNamespace)
	if err != nil {
		fmt.Println(err.Error())
		return ctrl.Result{Requeue: true}, err
	}

	for _, pod := range podlist.Items {

		// Build the command with arguments for podtracer

		podtracerArgs := []string{}
		podtracerArgs = append(podtracerArgs, "run")
		podtracerArgs = append(podtracerArgs, cmd.Spec.Command)
		podtracerArgs = append(podtracerArgs, "-a")
		podtracerArgs = append(podtracerArgs, cmd.Spec.Args)
		podtracerArgs = append(podtracerArgs, "--pod")
		podtracerArgs = append(podtracerArgs, pod.ObjectMeta.Name)
		podtracerArgs = append(podtracerArgs, "-n")
		podtracerArgs = append(podtracerArgs, pod.ObjectMeta.Namespace)

		// Temporarily listen to messages from podtracer on the operator pod
		// with nc and write those to file. Get the operator pod's IP and serve on port 5555 for now.

		// Generate the Cronjob object
		CronJob, err := r.CronJob(podtracerArgs, pod.ObjectMeta.Name, pod.Spec.NodeName, cmd.Spec.Schedule)
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
	}

	return ctrl.Result{}, nil
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
