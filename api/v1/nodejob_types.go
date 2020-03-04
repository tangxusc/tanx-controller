/*


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

package v1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// NodeJobSpec defines the desired state of NodeJob
type NodeJobSpec struct {
	PodSpec v1.PodSpec `json:"PodSpec"`
}

//nodename:podStatus
type NodePodStatus map[string]v1.PodPhase

// NodeJobStatus defines the observed state of NodeJob
type NodeJobStatus struct {
	Values NodePodStatus `json:"values,omitempty"`
	Finish bool          `json:"finish,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:printcolumn:name="images",type="string",JSONPath=".spec.PodSpec.containers[].image",description="images"
// +kubebuilder:printcolumn:name="status",type="string",JSONPath=".status.values",description="nodes status"
// +kubebuilder:printcolumn:name="finish",type="boolean",JSONPath=".status.finish",description="finish"

// NodeJob is the Schema for the nodejobs API
type NodeJob struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NodeJobSpec   `json:"spec,omitempty"`
	Status NodeJobStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// NodeJobList contains a list of NodeJob
type NodeJobList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NodeJob `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NodeJob{}, &NodeJobList{})
}
