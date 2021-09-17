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

package exec

import (
	"context"

	"fmt"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	execv1alpha1 "github.com/fennec-project/snoopy-operator/apis/exec/v1alpha1"
)

// CommandReconciler reconciles a Command object
type CommandReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=exec.fennecproject.io,resources=commands,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=exec.fennecproject.io,resources=commands/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=exec.fennecproject.io,resources=commands/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Command object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.9.2/pkg/reconcile
func (r *CommandReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	cmd := &execv1alpha1.Command{}
	err := r.Client.Get(ctx, req.NamespacedName, cmd)
	if err != nil {
		return ctrl.Result{}, err
	}

	podlist, err := r.GetRunningPodsByLabel(ctx, cmd.Spec.LabelSelector, cmd.Spec.TargetNamespace)
	if err != nil {
		fmt.Println(err.Error())
		return ctrl.Result{Requeue: true}, err
	}

	for _, pod := range podlist.Items {

		podtracerArgs := []string{}
		podtracerArgs = append(podtracerArgs, "run")
		podtracerArgs = append(podtracerArgs, cmd.Spec.Command)
		podtracerArgs = append(podtracerArgs, "-a")
		podtracerArgs = append(podtracerArgs, cmd.Spec.Args)
		podtracerArgs = append(podtracerArgs, "--pod")
		podtracerArgs = append(podtracerArgs, pod.ObjectMeta.Name)
		podtracerArgs = append(podtracerArgs, "-n")
		podtracerArgs = append(podtracerArgs, pod.ObjectMeta.Namespace)

		// Generate Command Job
		job, err := r.GenerateCommandJob(podtracerArgs, pod.ObjectMeta.Name, pod.Spec.NodeName)
		if err != nil {
			fmt.Println(err.Error())

			return ctrl.Result{}, err
		}
		// Create Job
		err = r.Client.Create(context.Background(), job)
		if err != nil {
			fmt.Println(err.Error())
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *CommandReconciler) GetRunningPodsByLabel(ctx context.Context, label map[string]string, namespace string) (*corev1.PodList, error) {

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

func (r *CommandReconciler) GenerateCommandJob(podtracerArgsList []string, targetPodName string, targetNodeName string) (*batchv1.Job, error) {

	var job *batchv1.Job

	var jobObjectMeta metav1.ObjectMeta
	var jobPodTemplate corev1.PodTemplateSpec
	var privileged bool
	var HostPathDirectory corev1.HostPathType
	var HostPathSocket corev1.HostPathType

	HostPathDirectory = "Directory"
	HostPathSocket = "Socket"

	privileged = true

	// TODO: improve the labeling system to identify jobs running
	jobObjectMeta = metav1.ObjectMeta{
		Name: "tcpdump-" + targetPodName,
		Labels: map[string]string{
			"tcpdumpJob": "snoopy-operator",
		},
		Namespace: "snoopy-operator",
	}

	jobPodTemplate = corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
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
						{Name: "kubeconfig",
							MountPath: "/root/.kube",
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

	job = &batchv1.Job{
		ObjectMeta: jobObjectMeta,
		Spec: batchv1.JobSpec{
			Template: jobPodTemplate,
		},
	}

	return job, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CommandReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&execv1alpha1.Command{}).
		Complete(r)
}
