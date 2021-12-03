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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type RawReceiver struct {
	Port     string `json:"port,omitempty"`
	FilePath string `json:"filePath,omitempty"`
}

// SnoopyDataEndpointSpec defines the desired state of SnoopyDataEndpoint
type SnoopyDataEndpointSpec struct {

	// ServiceName is used to create the service for the gRPC endpoint
	ServiceName string `json:"serviceName"`

	// ServicePort is the exposed port on the service
	ServicePort int32 `json:"servicePort"`
}

// SnoopyDataEndpointStatus defines the observed state of SnoopyDataEndpoint
type SnoopyDataEndpointStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// SnoopyDataEndpoint is the Schema for the snoopydataendpoints API
type SnoopyDataEndpoint struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SnoopyDataEndpointSpec   `json:"spec,omitempty"`
	Status SnoopyDataEndpointStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// SnoopyDataEndpointList contains a list of SnoopyDataEndpoint
type SnoopyDataEndpointList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SnoopyDataEndpoint `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SnoopyDataEndpoint{}, &SnoopyDataEndpointList{})
}
