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
	err := r.Client.Get(ctx, req.NamespacedName, tcpdump)
	if err != nil {
		return ctrl.Result{}, err
	}

	// Check tcpdump start time to see if it's already placed or in progress
	// TODO: refactor status field.
	if tcpdump.Status.StartTime != "" {
		fmt.Printf("tcpdump %s is running or already done.", tcpdump.ObjectMeta.Name)
		return ctrl.Result{}, nil
	}
	println(tcpdump.Spec.PodLabel)
	podlist, err := r.GetRunningPodsByLabel(ctx, tcpdump.Spec.PodLabel, tcpdump.Spec.TargetNamespace)
	if err != nil {
		fmt.Println(err.Error())
		return ctrl.Result{Requeue: true}, err
	}

	var tcpdumpArgs string
	for _, pod := range podlist.Items {

		// Generate tcpdump args with packetCount
		if tcpdump.Spec.PacketCount != 0 {

			tcpdumpArgs, err = r.GenerateTcpdumpArgs(tcpdump.Spec.InterfaceName, tcpdump.Spec.PacketCount, 0, tcpdump.Spec.PcapFilePath)
			if err != nil {
				fmt.Println(err.Error())
				return ctrl.Result{}, err
			}

		} else { // Generate tcpdump args with fileSize

			tcpdumpArgs, err = r.GenerateTcpdumpArgs(tcpdump.Spec.InterfaceName, 0, tcpdump.Spec.FileSize, tcpdump.Spec.PcapFilePath)
			if err != nil {
				fmt.Println(err.Error())
				return ctrl.Result{}, err
			}
		}

		podtracerArgs, err := r.GeneratePodtracerArgs("run", tcpdumpArgs, pod.ObjectMeta.Name, tcpdump.Spec.TargetNamespace)
		if err != nil {
			fmt.Println(err.Error())
			return ctrl.Result{}, err
		}

		// GenerateTcpdumpJob
		job, err := r.GenerateTcpdumpJob(podtracerArgs, pod.ObjectMeta.Name, pod.Spec.NodeName)
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
		// UpdateStatusOnCR

		// VerifyPcapFileAfterCompletion

	}

	return ctrl.Result{}, nil
}

func (r *TcpdumpReconciler) GetRunningPodsByLabel(ctx context.Context, label map[string]string, namespace string) (*corev1.PodList, error) {

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

func (r *TcpdumpReconciler) GeneratePodtracerArgs(command string, args string, targetPodName string, targetNamespace string) ([]string, error) {

	var podtracerArgsList []string

	podtracerArgsList = append(podtracerArgsList, command)
	podtracerArgsList = append(podtracerArgsList, "tcpdump")
	podtracerArgsList = append(podtracerArgsList, "-a")
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
		tcpdumpArgList = append(tcpdumpArgList, fmt.Sprint(packetCount))
		// tcpdumpArgList = append(tcpdumpArgList, "-w")
		// tcpdumpArgList = append(tcpdumpArgList, pcapFilePath)

	} else {
		tcpdumpArgList = append(tcpdumpArgList, "-i")
		tcpdumpArgList = append(tcpdumpArgList, interfaceName)
		tcpdumpArgList = append(tcpdumpArgList, "-C")
		tcpdumpArgList = append(tcpdumpArgList, fmt.Sprint(fileSize))
		// tcpdumpArgList = append(tcpdumpArgList, "-w")
		// tcpdumpArgList = append(tcpdumpArgList, pcapFilePath)
		tcpdumpArgList = append(tcpdumpArgList, "--pod")
	}

	tcpdumpArgs = strings.Join(tcpdumpArgList, " ")

	return tcpdumpArgs, nil
}

func (r *TcpdumpReconciler) GenerateTcpdumpJob(podtracerArgsList []string, targetPodName string, targetNodeName string) (*batchv1.Job, error) {

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
func (r *TcpdumpReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&pcapv1alpha1.Tcpdump{}).
		Complete(r)
}
