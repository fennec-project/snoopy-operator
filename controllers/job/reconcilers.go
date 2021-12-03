package job

import (
	"context"
	"fmt"

	jobv1alpha1 "github.com/fennec-project/snoopy-operator/apis/job/v1alpha1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	apimachinery "k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (r *SnoopyJobReconciler) reconcileCronJobs(snoopyJob *jobv1alpha1.SnoopyJob, cronJobs *batchv1.CronJobList) error {

	for _, cronJob := range cronJobs.Items {

		err := r.Client.Get(context.TODO(), apimachinery.NamespacedName{Namespace: cronJob.ObjectMeta.Namespace, Name: cronJob.ObjectMeta.Name}, &cronJob)
		if err != nil {
			if errors.IsNotFound(err) {
				err = r.Client.Create(context.Background(), &cronJob)
				if err != nil {
					fmt.Println(err.Error())
					return err
				}

				// Updating Status
				snoopyJob.Status.CronJobList = append(snoopyJob.Status.CronJobList, cronJob.ObjectMeta.Name)
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

	for _, job := range jobs.Items {

		err := r.Client.Get(context.TODO(), apimachinery.NamespacedName{Namespace: job.ObjectMeta.Namespace, Name: job.ObjectMeta.Name}, &job)
		if err != nil {
			if errors.IsNotFound(err) {

				err = r.Client.Create(context.Background(), &job)
				if err != nil {
					fmt.Println(err.Error())
					return err
				}

				// Updating Status
				snoopyJob.Status.CronJobList = append(snoopyJob.Status.CronJobList, job.ObjectMeta.Name)
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

	// Running reconciliation tasks
	// Target pod list by label and namespace
	podlist, err := r.getRunningPodsByLabel(context.TODO(), snoopyJob.Spec.LabelSelector, snoopyJob.Spec.TargetNamespace)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	cronJobs := &batchv1.CronJobList{}
	// CronJob creation by target pod
	for _, pod := range podlist.Items {

		// Build the command with arguments for podtracer
		podtracerOpts := r.buildPodtracerOptions(snoopyJob)

		podtracerOpts = append(podtracerOpts, "--pod")
		podtracerOpts = append(podtracerOpts, pod.ObjectMeta.Name)
		podtracerOpts = append(podtracerOpts, "-n")
		podtracerOpts = append(podtracerOpts, pod.ObjectMeta.Namespace)

		// Generate the Cronjob object
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

	// Running reconciliation tasks
	// Target pod list by label and namespace
	podlist, err := r.getRunningPodsByLabel(context.TODO(), snoopyJob.Spec.LabelSelector, snoopyJob.Spec.TargetNamespace)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	jobs := &batchv1.JobList{}
	// CronJob creation by target pod
	for _, pod := range podlist.Items {

		// Build the command with arguments for podtracer
		podtracerOpts := r.buildPodtracerOptions(snoopyJob)

		podtracerOpts = append(podtracerOpts, "--pod")
		podtracerOpts = append(podtracerOpts, pod.ObjectMeta.Name)
		podtracerOpts = append(podtracerOpts, "-n")
		podtracerOpts = append(podtracerOpts, pod.ObjectMeta.Namespace)

		// Generate the Cronjob object
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
