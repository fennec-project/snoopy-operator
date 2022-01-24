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

package data

import (
	"context"

	datav1alpha1 "github.com/fennec-project/snoopy-operator/apis/data/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// SnoopyDataEndpointReconciler reconciles a SnoopyDataEndpoint object
type SnoopyDataEndpointReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=data.fennecproject.io,resources=snoopydataendpoints,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=data.fennecproject.io,resources=snoopydataendpoints/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=data.fennecproject.io,resources=snoopydataendpoints/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the SnoopyDataEndpoint object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.9.2/pkg/reconcile
func (r *SnoopyDataEndpointReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	Log := log.FromContext(ctx).WithValues("method", "reconcile")

	Log.V(2).Info("Initiating reconciliation...")

	// get DataEndpoints
	DataEndpoint := &datav1alpha1.SnoopyDataEndpoint{}
	Log.V(2).Info("Looking for DataEndpoint CRs")

	err := r.Client.Get(ctx, req.NamespacedName, DataEndpoint)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue

			return reconcile.Result{}, nil
		}
		Log.Error(err, "Error requesting DataEndpoint")
		return ctrl.Result{Requeue: true}, err
	}

	// Reconcile Deployment for SnoopyDataEndpoint
	deploymentForDataEndpoint := &appsv1.Deployment{}
	objectMeta := setObjectMeta("snoopy-data", "snoopy-operator", map[string]string{"app": "snoopy-data"})
	err = r.reconcileResource(ctx, r.deploymentForDataEndpoint, DataEndpoint, deploymentForDataEndpoint, objectMeta)
	if err != nil {
		Log.Error(err, "Error reconciling deployment for SnoopyDataEndpoint...")
		return reconcile.Result{Requeue: true}, err
	}

	// Reconcile Service for SnoopyDataEndpoint
	svcForDataEndpoint := &corev1.Service{}
	objectMeta = setObjectMeta("snoopy-data-svc", "snoopy-operator", map[string]string{"app": "snoopy-data"})
	err = r.reconcileResource(ctx, r.serviceForDataEndpoint, DataEndpoint, svcForDataEndpoint, objectMeta)
	if err != nil {
		Log.Error(err, "Error reconciling deployment for SnoopyDataEndpoint...")
		return reconcile.Result{Requeue: true}, err
	}

	Log.V(2).Info("Reconciliation done successfully")
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *SnoopyDataEndpointReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&datav1alpha1.SnoopyDataEndpoint{}).
		Owns(&corev1.Service{}).
		Owns(&appsv1.Deployment{}).
		Complete(r)
}
