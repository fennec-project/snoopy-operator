package data

import (
	"context"
	"fmt"
	"time"

	datav1alpha1 "github.com/fennec-project/snoopy-operator/apis/data/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
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

func (r *SnoopyDataEndpointReconciler) reconcileResource(
	createResource createResourceFunc,
	dataEndpoint *datav1alpha1.SnoopyDataEndpoint,
	resource client.Object,
	objectMeta metav1.ObjectMeta) error {

	err := r.Client.Get(context.TODO(), types.NamespacedName{Name: objectMeta.Name, Namespace: objectMeta.Namespace}, resource)
	if err != nil {
		if errors.IsNotFound(err) {

			// Define a new resource

			fmt.Printf("\n%v - reconcileResource: creating a new resource for Snoopy Data Endpoint\n", time.Now())

			resource := createResource(dataEndpoint, objectMeta)
			err = r.Client.Create(context.TODO(), resource)

			if err != nil {
				fmt.Printf("\n%v - reconcileResource: failed to create new resource, err: %v\n", time.Now(), err)
				return err
			}

			// Resource created successfully - return and requeue
			return nil
		}

		fmt.Printf("\n%v - reconcileResource: error reading resource, err: %v\n", time.Now(), err)
		return err
	}

	return nil
}
