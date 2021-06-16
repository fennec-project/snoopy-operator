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
	"time"

	"strings"

	"os/exec"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	corev1 "k8s.io/api/core/v1"

	"github.com/containernetworking/plugins/pkg/ns"
	"github.com/vishvananda/netlink"

	pcapv1alpha1 "github.com/acmenezes/snoopy-operator/apis/pcap/v1alpha1"
)

// TcpdumpReconciler reconciles a Tcpdump object
type TcpdumpReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

var capBool bool

//+kubebuilder:rbac:groups=pcap.snoopy-operator.io,resources=tcpdumps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=pcap.snoopy-operator.io,resources=tcpdumps/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=pcap.snoopy-operator.io,resources=tcpdumps/finalizers,verbs=update

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

	targetPod := corev1.Pod{}
	err = r.Client.Get(context.Background(), client.ObjectKey{Namespace: "cnf-telco", Name: tcpdump.Spec.PodName}, &targetPod)
	if err != nil {
		return ctrl.Result{}, err
	}

	pid, err := getPid(targetPod)
	if err != nil {
		return ctrl.Result{}, err
	}

	// Get the pods namespace object
	targetNS, err := ns.GetNS("/host/proc/" + pid + "/ns/net")
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("Error getting Pod network namespace: %v", err)
	}

	err = targetNS.Do(func(hostNs ns.NetNS) error {

		_, err := netlink.LinkByName(tcpdump.Spec.IfName)
		if err != nil {
			return fmt.Errorf("Interface could not be found: %v", err)
		}

		pcapFile := "/pcap-data/" + tcpdump.Spec.PodName + "_" + strings.ReplaceAll(time.Now().String(), " ", "_") + ".pcap"

		// Running tcpdump on given Pod and Interface
		cmd := exec.Command("tcpdump", "-i", tcpdump.Spec.IfName, "-w", pcapFile)

		if err := cmd.Start(); err != nil {
			return err
		}

		// Updating start time for the current tcpdump packet capture
		tcpdump.Status.StartTime = time.Now().String()
		tcpdump.Status.PcapFilePath = pcapFile

		err = r.Client.Update(context.Background(), tcpdump)
		if err != nil {
			return err
		}

		fmt.Printf("Starting tcpdump on interface %s at %v\n", tcpdump.Spec.IfName, time.Now())

		time.Sleep(time.Duration(tcpdump.Spec.Duration) * time.Minute)

		if err := cmd.Process.Kill(); err != nil {
			return err
		}

		fmt.Printf("Stopping tcpdump on interface %s at %v\n", tcpdump.Spec.IfName, time.Now())

		// Updating end time for the current tcpdump packet capture
		tcpdump.Status.EndTime = time.Now().String()
		err = r.Client.Update(context.Background(), tcpdump)
		if err != nil {
			return err
		}

		return nil
	})

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *TcpdumpReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&pcapv1alpha1.Tcpdump{}).
		Complete(r)
}
