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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// TcpdumpSpec defines the desired state of Tcpdump
type TcpdumpSpec struct {

	// Target pod where the tcpdump command will get packets from
	PodName string `json:"podName,omitempty"`

	// Network virtual interface or link name on the pod to run tcpdump on
	IfName string `json:"ifName,omitempty"`
}

// TcpdumpStatus defines the observed state of Tcpdump
type TcpdumpStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Tcpdump is the Schema for the tcpdumps API
type Tcpdump struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TcpdumpSpec   `json:"spec,omitempty"`
	Status TcpdumpStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// TcpdumpList contains a list of Tcpdump
type TcpdumpList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Tcpdump `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Tcpdump{}, &TcpdumpList{})
}
