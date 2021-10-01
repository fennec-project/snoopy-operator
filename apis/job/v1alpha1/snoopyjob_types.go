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

// SnoopyJobSpec defines the desired state of SnoopyJob
type SnoopyJobSpec struct {
	// Command is any linux binary that can be run by podtracer in the context of a Pod
	// Warning: The command must be present in the used potracer image for it to be used
	Command string `json:"command,omitempty"`

	// Args is a string containing all arguments for a given command
	Args string `json:"args,omitempty"`

	// LabelSelector is the label to find the target Pods
	LabelSelector map[string]string `json:"labelSelector,omitempty"`

	// TargetNamespace is the k8s where the target Pod lives
	TargetNamespace string `json:"targetNamespace,omitempty"`

	// The schedule in Cron format, see https://en.wikipedia.org/wiki/Cron.
	Schedule string `json:"schedule,omitempty"`
}

// SnoopyJobStatus defines the observed state of SnoopyJob
type SnoopyJobStatus struct {
	CronJobList []string `json:"cronJobList,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// SnoopyJob is the Schema for the snoopyjobs API
type SnoopyJob struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SnoopyJobSpec   `json:"spec,omitempty"`
	Status SnoopyJobStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// SnoopyJobList contains a list of SnoopyJob
type SnoopyJobList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SnoopyJob `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SnoopyJob{}, &SnoopyJobList{})
}
