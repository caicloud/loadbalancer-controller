/*
Copyright 2018 Caicloud.

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

// SpringcloudSpec defines the desired state of Springcloud
type SpringcloudSpec struct {
	// Description is the description of current microservices
	Description string `json:"description,omitempty"`
	// MicroserviceClass is the type of microservices, e.g. springcloud
	MicroserviceClass MicroserviceClass `json:"microserviceClass"`
	// Releases is the map of releases. key ==> release's name, value ==> alias
	Releases map[string]string `json:"releases"`
}

// SpringcloudStatus defines the observed state of Springcloud
type SpringcloudStatus struct {
	Phase          SpringcloudPhase `json:"phase,omitempty"`
	Message        string           `json:"message,omitempty"`
	LastUpdateTime *metav1.Time     `json:"lastUpdateTime,omitempty"`
}

// SpringcloudPhase is a label for the condition of a microservices at the current time.
type SpringcloudPhase string

const (
	// ResourceDeploying means microsvc is deploying
	ResourceDeploying SpringcloudPhase = "Deploying"
	// ResourceRunning means micsrosvc is running
	ResourceRunning SpringcloudPhase = "Running"
	// ResourceFailed means microsvc is failed
	ResourceFailed SpringcloudPhase = "Failed"
	// ResourceTerminating means microsvc is terminating
	ResourceTerminating SpringcloudPhase = "Terminating"
)

// MicroserviceClass define the class of a microservice.
type MicroserviceClass string

const (
	// SpringcloudClass means that the microservice is using springCloud.
	SpringcloudClass MicroserviceClass = "springcloud"
)

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Springcloud is the Schema for the microservicess API
type Springcloud struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SpringcloudSpec   `json:"spec,omitempty"`
	Status SpringcloudStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SpringcloudList contains a list of Springcloud
type SpringcloudList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Springcloud `json:"items"`
}
