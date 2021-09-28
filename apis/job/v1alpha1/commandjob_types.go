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

// CommandJobSpec defines the desired state of CommandJob
type CommandJobSpec struct {
	// Command is any linux binary that can be run by podtracer in the context of a Pod
	// Warning: The command must be present in the used potracer image for it to be used
	Command string `json:"command,omitempty"`

	// Args is a string containing all arguments for a given command
	Args string `json:"args,omitempty"`

	// LabelSelector is the label to find the target Pods
	LabelSelector map[string]string `json:"labelSelector,omitemtpy"`

	// TargetNamespace is the k8s where the target Pod lives
	TargetNamespace string `json:"targetNamespace,omitempty"`
}

// CommandJobStatus defines the observed state of CommandJob
type CommandJobStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// CommandJob is the Schema for the commandjobs API
type CommandJob struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CommandJobSpec   `json:"spec,omitempty"`
	Status CommandJobStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// CommandJobList contains a list of CommandJob
type CommandJobList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CommandJob `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CommandJob{}, &CommandJobList{})
}
