package job

import (
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (r *CommandJobReconciler) CronJob(podtracerArgsList []string, targetPodName string, targetNodeName string, schedule string) (*batchv1.CronJob, error) {

	var CronJob *batchv1.CronJob
	var privileged bool
	var HostPathDirectory corev1.HostPathType
	var HostPathSocket corev1.HostPathType

	HostPathDirectory = "Directory"
	HostPathSocket = "Socket"

	privileged = true

	// TODO: improve the labeling system to identify jobs running
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
						{Name: "pcap-data",
							MountPath: "/pcap-data",
							ReadOnly:  false},
						// {Name: "kubeconfig",
						// 	MountPath: "/root/.kube",
						// 	ReadOnly:  false},
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
				{
					Name: "pcap-data",
					VolumeSource: corev1.VolumeSource{
						EmptyDir: &corev1.EmptyDirVolumeSource{},
					},
				},
				{
					Name: "kubeconfig",
					VolumeSource: corev1.VolumeSource{
						Secret: &corev1.SecretVolumeSource{
							SecretName: "podtracer-kubeconfig",
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
				"snoopyJob": "CommandJob",
			},
			Namespace: "snoopy-operator",
		},
		Spec: JobSpec,
	}

	// CronJobSpec vars
	var StartingDeadlineSeconds *int64
	ConcurrencyPolicy := batchv1.ReplaceConcurrent
	var Suspend *bool
	var SuccessfulJobsHistoryLimit *int32
	var FailedJobsHistoryLimit *int32

	CronJob = &batchv1.CronJob{

		ObjectMeta: metav1.ObjectMeta{
			Name: "snoopy-cronjob-" + targetPodName,
			Labels: map[string]string{
				"snoopyCronJob": "CommandJob",
			},
			Namespace: "snoopy-operator",
		},

		Spec: batchv1.CronJobSpec{
			// The schedule in Cron format, see https://en.wikipedia.org/wiki/Cron.
			Schedule: schedule,

			// Optional deadline in seconds for starting the job if it misses scheduled
			// time for any reason.  Missed jobs executions will be counted as failed ones.
			// optional
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
			JobTemplate: JobTemplateSpec,

			// The number of successful finished jobs to retain.
			// This is a pointer to distinguish between explicit zero and not specified.
			// optional
			SuccessfulJobsHistoryLimit: SuccessfulJobsHistoryLimit,

			// The number of failed finished jobs to retain.
			// This is a pointer to distinguish between explicit zero and not specified.
			// optional
			FailedJobsHistoryLimit: FailedJobsHistoryLimit,
		},
	}

	return CronJob, nil
}
