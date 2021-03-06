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

	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	datav1alpha1 "github.com/fennec-project/snoopy-operator/apis/data/v1alpha1"
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
	// Set dataEndpoint instance as the owner and controller.
	if err := controllerutil.SetControllerReference(dataEndpoint, service, r.Scheme); err != nil {
		log.Fatal(err)
	}
	return service
}
