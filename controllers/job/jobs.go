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
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (r *SnoopyJobReconciler) Job(podtracerArgsList []string, targetPodName string, targetNodeName string) (*batchv1.Job, error) {

	jobTemplateSpec, err := r.JobTemplateSpec(podtracerArgsList, targetPodName, targetNodeName)
	if err != nil {
		return nil, err
	}

	job := &batchv1.Job{
		ObjectMeta: jobTemplateSpec.ObjectMeta,
		Spec:       jobTemplateSpec.Spec,
	}

	return job, nil
}

func (r *SnoopyJobReconciler) CronJob(podtracerArgsList []string, targetPodName string, targetNodeName string, schedule string) (*batchv1.CronJob, error) {

	var CronJob *batchv1.CronJob

	// CronJobSpec vars
	var StartingDeadlineSeconds *int64
	ConcurrencyPolicy := batchv1.ReplaceConcurrent
	var Suspend *bool
	var SuccessfulJobsHistoryLimit *int32
	var FailedJobsHistoryLimit *int32

	jobTemplateSpec, err := r.JobTemplateSpec(podtracerArgsList, targetPodName, targetNodeName)
	if err != nil {
		return nil, err
	}

	CronJob = &batchv1.CronJob{

		ObjectMeta: metav1.ObjectMeta{
			Name: "snoopy-cronjob-" + targetPodName,
			Labels: map[string]string{
				"snoopyCronJob": "SnoopyJob",
			},
			Namespace: "snoopy-operator",
		},

		Spec: batchv1.CronJobSpec{
			// The schedule in Cron format, see https://en.wikipedia.org/wiki/Cron.
			Schedule: schedule,

			// Optional deadline in seconds for starting the job if it misses scheduled.
			// time for any reason.  Missed jobs executions will be counted as failed ones.
			// optional.
			StartingDeadlineSeconds: StartingDeadlineSeconds,

			// Specifies how to treat concurrent executions of a Job.
			// Valid values are:
			// - "Allow" (default): allows CronJobs to run concurrently;
			// - "Forbid": forbids concurrent runs, skipping next run if previous run hasn't finished yet;
			// - "Replace": cancels currently running job and replaces it with a new one
			// optional
			ConcurrencyPolicy: ConcurrencyPolicy,

			// This flag tells the controller to suspend subsequent executions, it does
			// not apply to already started executions. Defaults to false.
			// optional
			Suspend: Suspend,

			// Specifies the job that will be created when executing a CronJob.
			JobTemplate: *jobTemplateSpec,

			// The number of successful finished jobs to retain.
			// This is a pointer to distinguish between explicit zero and not specified.
			// optional.
			SuccessfulJobsHistoryLimit: SuccessfulJobsHistoryLimit,

			// The number of failed finished jobs to retain.
			// This is a pointer to distinguish between explicit zero and not specified.
			// optional.
			FailedJobsHistoryLimit: FailedJobsHistoryLimit,
		},
	}
	return CronJob, nil
}

func (r *SnoopyJobReconciler) JobTemplateSpec(podtracerArgsList []string, targetPodName string, targetNodeName string) (*batchv1.JobTemplateSpec, error) {
	var privileged bool
	var HostPathDirectory corev1.HostPathType
	var HostPathSocket corev1.HostPathType

	HostPathDirectory = "Directory"
	HostPathSocket = "Socket"

	privileged = true

	// TODO: improve the labeling system to identify jobs running.
	PodTemplateSpec := corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "snoopy-worker",
			Labels: map[string]string{"app": "go-remote"},
		},
		Spec: corev1.PodSpec{
			NodeName:           targetNodeName,
			ServiceAccountName: serviceAccountName,
			RestartPolicy:      "Never",
			Containers: []corev1.Container{
				{
					Name:            "podtracer",
					Image:           podtracerImage,
					ImagePullPolicy: corev1.PullAlways,
					Command:         []string{"/usr/bin/podtracer"},
					Args:            podtracerArgsList,
					SecurityContext: &corev1.SecurityContext{
						Privileged: &privileged,
					},
					VolumeMounts: []corev1.VolumeMount{
						{Name: "proc",
							MountPath: "/host/proc",
							ReadOnly:  false},
						{Name: "crio-sock",
							MountPath: "/var/run/crio/crio.sock",
							ReadOnly:  false},
					},
				},
			},
			Volumes: []corev1.Volume{{
				Name: "proc",
				VolumeSource: corev1.VolumeSource{
					HostPath: &corev1.HostPathVolumeSource{
						Path: "/proc",
						Type: &HostPathDirectory,
					},
				},
			},
				{
					Name: "crio-sock",
					VolumeSource: corev1.VolumeSource{
						HostPath: &corev1.HostPathVolumeSource{
							Path: "/var/run/crio/crio.sock",
							Type: &HostPathSocket,
						},
					},
				},
			},
		},
	}

	JobSpec := batchv1.JobSpec{
		Template: PodTemplateSpec,
	}

	JobTemplateSpec := batchv1.JobTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Name: "snoopy-job-" + targetPodName,
			Labels: map[string]string{
				"snoopyJob": "SnoopyJob",
			},
			Namespace: "snoopy-operator",
		},
		Spec: JobSpec,
	}

	return &JobTemplateSpec, nil
}
