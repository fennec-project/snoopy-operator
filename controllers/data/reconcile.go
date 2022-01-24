package data

import (
	"context"

	datav1alpha1 "github.com/fennec-project/snoopy-operator/apis/data/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type createResourceFunc func(dataEndpoint *datav1alpha1.SnoopyDataEndpoint, objectMeta metav1.ObjectMeta) client.Object

func setObjectMeta(name string, namespace string, labels map[string]string) metav1.ObjectMeta {
	objectMeta := metav1.ObjectMeta{
		Name:      name,
		Namespace: namespace,
		Labels:    labels,
	}
	return objectMeta
}

func (r *SnoopyDataEndpointReconciler) reconcileResource(ctx context.Context,
	createResource createResourceFunc,
	dataEndpoint *datav1alpha1.SnoopyDataEndpoint,
	resource client.Object,
	objectMeta metav1.ObjectMeta) error {

	Log := log.FromContext(ctx).WithValues("method", "reconcileResource")

	err := r.Client.Get(context.TODO(), types.NamespacedName{Name: objectMeta.Name, Namespace: objectMeta.Namespace}, resource)
	if err != nil {
		if errors.IsNotFound(err) {

			// Define a new resource
			Log.Info("Creating a new resource for Snoopy Data Endpoint")

			resource := createResource(dataEndpoint, objectMeta)
			err = r.Client.Create(context.TODO(), resource)

			if err != nil {
				Log.Error(err, "Failed to create new resource")
				return err
			}

			Log.Info("Resource created successfully", "resource", resource.GetName())
			return nil
		}

		Log.Error(err, "Error reading resource", "resource", resource.GetName())
		return err
	}

	return nil
}
