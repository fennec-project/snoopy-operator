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

	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	corev1 "k8s.io/api/core/v1"

	pcapv1alpha1 "github.com/fennec-project/snoopy-operator/apis/pcap/v1alpha1"
)

// TcpdumpReconciler reconciles a Tcpdump object
type TcpdumpReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

var capBool bool

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

	if tcpdump.Status.StartTime != "" {
		fmt.Printf("tcpdump %s is running or already done.", tcpdump.ObjectMeta.Name)
		return ctrl.Result{}, nil
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

// SetupWithManager sets up the controller with the Manager.
func (r *TcpdumpReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&pcapv1alpha1.Tcpdump{}).
		Complete(r)
}
