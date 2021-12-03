package data

import (
	datav1alpha1 "github.com/fennec-project/snoopy-operator/apis/data/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func (r *SnoopyDataEndpointReconciler) serviceForDataEndpoint(dataEndpoint *datav1alpha1.SnoopyDataEndpoint, objectMeta metav1.ObjectMeta) client.Object {

	service := &corev1.Service{
		ObjectMeta: objectMeta,
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{Name: dataEndpoint.Spec.ServiceName,
					Protocol:   "TCP",
					Port:       dataEndpoint.Spec.ServicePort,
					TargetPort: intstr.FromInt(51001)},
			},
			Selector: objectMeta.Labels,
		},
	}
	// Set dataEndpoint instance as the owner and controller
	controllerutil.SetControllerReference(dataEndpoint, service, r.Scheme)
	return service
}
