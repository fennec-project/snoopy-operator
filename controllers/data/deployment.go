// Copyright The Snoopy Operator Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package data

import (
	"log"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	datav1alpha1 "github.com/fennec-project/snoopy-operator/apis/data/v1alpha1"
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
	// Set dataEndpoint instance as the owner and controller.
	if err := controllerutil.SetControllerReference(dataEndpoint, deploy, r.Scheme); err != nil {
		log.Fatal(err)
	}

	return deploy
}
