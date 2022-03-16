// Copyright The Snoopy Operator Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package job

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apimachinery "k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"

	jobv1alpha1 "github.com/fennec-project/snoopy-operator/apis/job/v1alpha1"
)

func (r *SnoopyJobReconciler) reconcileCronJobs(snoopyJob *jobv1alpha1.SnoopyJob, cronJobs *batchv1.CronJobList) error {

	for i := range cronJobs.Items {

		err := r.Client.Get(context.TODO(), apimachinery.NamespacedName{Namespace: cronJobs.Items[i].ObjectMeta.Namespace, Name: cronJobs.Items[i].ObjectMeta.Name}, &cronJobs.Items[i])
		if err != nil {
			if errors.IsNotFound(err) {
				err = r.Client.Create(context.Background(), &cronJobs.Items[i])
				if err != nil {
					fmt.Println(err.Error())
					return err
				}

				// Updating Status.
				snoopyJob.Status.CronJobList = append(snoopyJob.Status.CronJobList, cronJobs.Items[i].ObjectMeta.Name)
				err = r.Client.Status().Update(context.Background(), snoopyJob)
				if err != nil {
					return nil
				}
			} else {

				return err
			}
		}
	}

	return nil
}

func (r *SnoopyJobReconciler) reconcileJobs(snoopyJob *jobv1alpha1.SnoopyJob, jobs *batchv1.JobList) error {

	for i := range jobs.Items {

		err := r.Client.Get(context.TODO(), apimachinery.NamespacedName{Namespace: jobs.Items[i].ObjectMeta.Namespace, Name: jobs.Items[i].ObjectMeta.Name}, &jobs.Items[i])
		if err != nil {
			if errors.IsNotFound(err) {

				err = r.Client.Create(context.Background(), &jobs.Items[i])
				if err != nil {
					fmt.Println(err.Error())
					return err
				}

				// Updating Status.
				snoopyJob.Status.CronJobList = append(snoopyJob.Status.CronJobList, jobs.Items[i].ObjectMeta.Name)
				err = r.Client.Status().Update(context.Background(), snoopyJob)
				if err != nil {
					return err
				}

			} else {
				return err
			}
		}
	}

	return nil
}

func (r *SnoopyJobReconciler) buildCronJobForPods(snoopyJob *jobv1alpha1.SnoopyJob) (*batchv1.CronJobList, error) {

	// Running reconciliation tasks.
	// Target pod list by label and namespace.
	podlist, err := r.getRunningPodsByLabel(context.TODO(), snoopyJob.Spec.LabelSelector, snoopyJob.Spec.TargetNamespace)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	cronJobs := &batchv1.CronJobList{}
	// CronJob creation by target pod.
	for _, pod := range podlist.Items {

		// Build the command with arguments for podtracer.
		podtracerOpts := r.buildPodtracerOptions(snoopyJob)

		podtracerOpts = append(podtracerOpts, "--pod")
		podtracerOpts = append(podtracerOpts, pod.ObjectMeta.Name)
		podtracerOpts = append(podtracerOpts, "-n")
		podtracerOpts = append(podtracerOpts, pod.ObjectMeta.Namespace)

		// Generate the Cronjob object.
		cronJob, err := r.CronJob(podtracerOpts, pod.ObjectMeta.Name, pod.Spec.NodeName, snoopyJob.Spec.Schedule)
		if err != nil {
			return nil, err
		}
		if err := ctrl.SetControllerReference(snoopyJob, cronJob, r.Scheme); err != nil {
			return nil, err
		}

		cronJobs.Items = append(cronJobs.Items, *cronJob)
	}
	return cronJobs, nil
}

func (r *SnoopyJobReconciler) buildJobForPods(snoopyJob *jobv1alpha1.SnoopyJob) (*batchv1.JobList, error) {

	// Running reconciliation tasks.
	// Target pod list by label and namespace.
	podlist, err := r.getRunningPodsByLabel(context.TODO(), snoopyJob.Spec.LabelSelector, snoopyJob.Spec.TargetNamespace)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	jobs := &batchv1.JobList{}
	// CronJob creation by target pod.
	for _, pod := range podlist.Items {

		// Build the command with arguments for podtracer.
		podtracerOpts := r.buildPodtracerOptions(snoopyJob)

		podtracerOpts = append(podtracerOpts, "--pod")
		podtracerOpts = append(podtracerOpts, pod.ObjectMeta.Name)
		podtracerOpts = append(podtracerOpts, "-n")
		podtracerOpts = append(podtracerOpts, pod.ObjectMeta.Namespace)

		// Generate the Cronjob object.
		job, err := r.Job(podtracerOpts, pod.ObjectMeta.Name, pod.Spec.NodeName)
		if err != nil {
			return nil, err
		}
		if err := ctrl.SetControllerReference(snoopyJob, job, r.Scheme); err != nil {
			return nil, err
		}

		jobs.Items = append(jobs.Items, *job)
	}
	return jobs, nil
}

func (r *SnoopyJobReconciler) buildPodtracerOptions(snoopyJob *jobv1alpha1.SnoopyJob) []string {

	podtracerOpts := []string{}
	podtracerOpts = append(podtracerOpts, "run")
	podtracerOpts = append(podtracerOpts, snoopyJob.Spec.Command)
	podtracerOpts = append(podtracerOpts, "-a")
	podtracerOpts = append(podtracerOpts, snoopyJob.Spec.Args)

	if snoopyJob.Spec.Timer != "" {
		podtracerOpts = append(podtracerOpts, "-t")
		podtracerOpts = append(podtracerOpts, snoopyJob.Spec.Timer)
	}

	if snoopyJob.Spec.DataServiceIP != "" {
		podtracerOpts = append(podtracerOpts, "-d")
		podtracerOpts = append(podtracerOpts, snoopyJob.Spec.DataServiceIP)
		podtracerOpts = append(podtracerOpts, "-p")
		podtracerOpts = append(podtracerOpts, snoopyJob.Spec.DataServicePort)
	}

	return podtracerOpts
}

func (r *SnoopyJobReconciler) getRunningPodsByLabel(ctx context.Context, label map[string]string, namespace string) (*corev1.PodList, error) {

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
