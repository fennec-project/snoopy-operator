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

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	datav1alpha1 "github.com/fennec-project/snoopy-operator/apis/data/v1alpha1"

	zap "go.uber.org/zap"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// SnoopyDataEndpointReconciler reconciles a SnoopyDataEndpoint object
type SnoopyDataEndpointReconciler struct {
	client.Client
	Scheme       *runtime.Scheme
	DataEndpoint *datav1alpha1.SnoopyDataEndpoint
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
	_ = log.FromContext(ctx)

	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// Log User Info message about new jobs
	logger.Info("Checking for new Snoopy Data Endpoints")

	// get DataEndpoints
	r.DataEndpoint = &datav1alpha1.SnoopyDataEndpoint{}
	err := r.Client.Get(ctx, req.NamespacedName, r.DataEndpoint)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// Check for deletion timestamp and finalizers
	finalizer := "snoopydataendpoint.job.fennecproject.io"
	if r.DataEndpoint.ObjectMeta.DeletionTimestamp.IsZero() {

		// DataEndpoint is not being deleted, so if it does not have our finalizer,
		// then lets add the finalizer and update the object. This is equivalent
		// registering our finalizer.

		if !containsString(r.DataEndpoint.GetFinalizers(), finalizer) {

			logger.Info("New DataEndpoint found setting finalizers")

			r.DataEndpoint.SetFinalizers(append(r.DataEndpoint.GetFinalizers(), finalizer))
			if err := r.Update(context.Background(), r.DataEndpoint); err != nil {
				return ctrl.Result{Requeue: true}, err
			}
			// Running reconciliation tasks

		}

	} else {
		// DataEndpoint is being deleted
		if containsString(r.DataEndpoint.GetFinalizers(), finalizer) {

			// remove our finalizer from the list and update it.
			r.DataEndpoint.SetFinalizers(removeString(r.DataEndpoint.GetFinalizers(), finalizer))
			if err := r.Update(context.Background(), r.DataEndpoint); err != nil {
				return ctrl.Result{Requeue: true}, err
			}
		}
	}

	// Log User Info message when job is being deleted or registering finalizer

	// List Target Pods matching labels on selected namespace

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *SnoopyDataEndpointReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&datav1alpha1.SnoopyDataEndpoint{}).
		Complete(r)
}

// Helper functions to check and remove string from a slice of strings.
func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func removeString(slice []string, s string) (result []string) {
	for _, item := range slice {
		if item == s {
			continue
		}
		result = append(result, item)
	}
	return
}
