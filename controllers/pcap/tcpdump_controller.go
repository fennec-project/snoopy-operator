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

package pcap

import (
	"context"
	"strings"

	"fmt"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	pcapv1alpha1 "github.com/fennec-project/snoopy-operator/apis/pcap/v1alpha1"
)

// TcpdumpReconciler reconciles a Tcpdump object
type TcpdumpReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=pcap.fennecproject.io,resources=tcpdumps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=pcap.fennecproject.io,resources=tcpdumps/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=pcap.fennecproject.io,resources=tcpdumps/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *TcpdumpReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	tcpdump := &pcapv1alpha1.Tcpdump{}
	err := r.Client.Get(context.Background(), req.NamespacedName, tcpdump)
	if err != nil {
		return ctrl.Result{}, err
	}

	// Check tcpdump start time to see if it's already placed or in progress
	// TODO: refactor status field.
	if tcpdump.Status.StartTime != "" {
		fmt.Printf("tcpdump %s is running or already done.", tcpdump.ObjectMeta.Name)
		return ctrl.Result{}, nil
	}

	podlist, err := r.GetRunningPodsByLabel(tcpdump.Spec.PodLabel, tcpdump.Spec.TargetNamespace)
	if err != nil {
		fmt.Printf(err.Error())
		return ctrl.Result{Requeue: true}, err
	}

	for _, pod := range podlist.Items {

		// GeneratePodtracerArgs with packetCount
		if tcpdump.Spec.PacketCount != 0 {

			podtracerArgs, err := r.GeneratePodtracerArgs(tcpdump.Spec.InterfaceName, tcpdump.Spec.PacketCount, 0, tcpdump.Spec.PcapFilePath, pod.ObjectMeta.Name, tcpdump.Spec.TargetNamespace)
			if err != nil {
				fmt.Printf(err.Error())
				return ctrl.Result{}, err
			}

		} else { // GeneratePodtracerArgs with fileSize

			podtracerArgs, err = r.GeneratePodtracerArgs(tcpdump.Spec.InterfaceName, 0, tcpdump.Spec.FileSize, tcpdump.Spec.PcapFilePath, pod.ObjectMeta.Name, tcpdump.Spec.TargetNamespace)
			if err != nil {
				fmt.Printf(err.Error())
				return ctrl.Result{}, err
			}
		}

		// GenerateTcpdumpJob

		// UpdateStatusOnCR

		// VerifyPcapFileAfterCompletion

	}

	return ctrl.Result{}, nil
}

func (r *TcpdumpReconciler) GetRunningPodsByLabel(label map[string]string, namespace string) (*corev1.PodList, error) {

	var podlist *corev1.PodList
	err := r.Client.List(context.Background(), podlist, client.MatchingLabels(label), client.InNamespace(namespace), client.MatchingFields{"Status.Phase": "Running"})
	if err != nil {
		fmt.Printf("GetRunningPodsByLabel, Error listing pods for tcpdump %s ", err.Error())
		return nil, err
	}

	if len(podlist.Items) <= 0 {
		return nil, fmt.Errorf("no running pod corresponds to label %v and namespace %v ", label, namespace)
	}

	return podlist, nil
}

func (r *TcpdumpReconciler) GeneratePodtracerArgs(command string, args string, targetPodName string, targetNamespace string) ([]string, error) {

	var podtracerArgsList []string

	podtracerArgsList = append(podtracerArgsList, command)
	podtracerArgsList = append(podtracerArgsList, args)
	podtracerArgsList = append(podtracerArgsList, "--pod")
	podtracerArgsList = append(podtracerArgsList, targetPodName)
	podtracerArgsList = append(podtracerArgsList, "-n")
	podtracerArgsList = append(podtracerArgsList, targetNamespace)

	return podtracerArgsList, nil
}

func (r *TcpdumpReconciler) GenerateTcpdumpArgs(interfaceName string, packetCount int64, fileSize int64, pcapFilePath string) (string, error) {

	var tcpdumpArgs string
	var tcpdumpArgList []string

	if packetCount != 0 {
		tcpdumpArgList = append(tcpdumpArgList, "-i")
		tcpdumpArgList = append(tcpdumpArgList, interfaceName)
		tcpdumpArgList = append(tcpdumpArgList, "-c")
		tcpdumpArgList = append(tcpdumpArgList, string(packetCount))
		tcpdumpArgList = append(tcpdumpArgList, "-w")
		tcpdumpArgList = append(tcpdumpArgList, pcapFilePath)

	} else {
		tcpdumpArgList = append(tcpdumpArgList, "-i")
		tcpdumpArgList = append(tcpdumpArgList, interfaceName)
		tcpdumpArgList = append(tcpdumpArgList, "-C")
		tcpdumpArgList = append(tcpdumpArgList, string(fileSize))
		tcpdumpArgList = append(tcpdumpArgList, "-w")
		tcpdumpArgList = append(tcpdumpArgList, pcapFilePath)
		tcpdumpArgList = append(tcpdumpArgList, "--pod")
	}

	tcpdumpArgs = strings.Join(tcpdumpArgList, " ")

	return tcpdumpArgs, nil
}

func (r *TcpdumpReconciler) GenerateTcpdumpJob(podtracerArgsList []string, targetPodName string) (*batchv1.Job, error) {

	var job *batchv1.Job

	var jobObjectMeta metav1.ObjectMeta
	var jobPodTemplate corev1.PodTemplateSpec
	var jobSpec batchv1.JobSpec
	var privileged bool

	privileged = true

	// TODO: improve the labeling system to identify jobs running
	jobObjectMeta = metav1.ObjectMeta{
		Name: "tcpdumpJobForPod-" + targetPodName,
		Labels: map[string]string{
			"tcpdumpJob": "snoopy-operator",
		},
	}

	jobPodTemplate = corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{"app": "go-remote"},
		},
		Spec: corev1.PodSpec{
			ServiceAccountName: serviceAccountName,
			Containers: []corev1.Container{
				{
					Name:            "podtracer",
					Image:           podtracerImage,
					ImagePullPolicy: corev1.PullAlways,
					Command:         []string{"podtracer"},
					Args:            podtracerArgsList,
					SecurityContext: &corev1.SecurityContext{
						Privileged: &privileged,
					},
					VolumeMounts: goRemote.Spec.VolumeMounts,
				},
			},
			Volumes: goRemote.Spec.Volumes,
		},
	}

	jobSpec = batchv1.JobSpec{}

	// err = r.Client.Create(context.Background(), job)

	return job, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *TcpdumpReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&pcapv1alpha1.Tcpdump{}).
		Complete(r)
}
