package data

import (
	datav1alpha1 "github.com/fennec-project/snoopy-operator/apis/data/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func (r *SnoopyDataEndpointReconciler) deploymentForDataEndpoint(dataEndpoint *datav1alpha1.SnoopyDataEndpoint, objectMeta metav1.ObjectMeta) client.Object {

	privmode := true

	deploy := &appsv1.Deployment{
		ObjectMeta: objectMeta,
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: objectMeta.Labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: objectMeta.Labels,
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: serviceAccountName,
					Containers: []corev1.Container{{
						Name:            objectMeta.Name,
						Image:           dataEndpointImage,
						ImagePullPolicy: corev1.PullAlways,
						Command:         []string{"/server"},
						Args:            []string{"51001"},
						SecurityContext: &corev1.SecurityContext{
							Privileged: &privmode,
						},
						Ports: []corev1.ContainerPort{{
							ContainerPort: 51001,
						}},
					}},
				},
			},
		},
	}
	// Set dataEndpoint instance as the owner and controller
	controllerutil.SetControllerReference(dataEndpoint, deploy, r.Scheme)
	return deploy
}
