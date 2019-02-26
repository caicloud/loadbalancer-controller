package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// LogEndpointList describes an array of log endpoint instances.
type LogEndpointList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []LogEndpoint `json:"items"`
}

// +genclient
// +genclient:nonNamespaced
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// LogEndpoint describes an instance of log endpoint.
type LogEndpoint struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              LogEndpointSpec   `json:"spec"`
	Status            LogEndpointStatus `json:"status,omitempty"`
}

// LogEndpointSpec describes specification of a log endpoint instance.
type LogEndpointSpec struct {
	Description       string `json:"description,omitempty"`
	Target            `json:",inline"`
	Default           bool     `json:"default,omitempty"`
	CollectedClusters []string `json:"collectedClusters,omitempty"`
}

// Target represents the target to output logs.
// Only support one type of endpoint at a time.
type Target struct {
	Kafka         *KafkaEndpoint         `json:"kafka,omitempty"`
	Elasticsearch *ElasticsearchEndpoint `json:"elasticsearch,omitempty"`
}

// KafkaEndpoint represents the Kafka endpoint to output logs.
type KafkaEndpoint struct {
	Brokers []string `json:"brokers"`
	Topic   string   `json:"topic"`
	Version string   `json:"version"`
	// Consumers represents the array of Elasticsearch to consume logs from Kafka.
	Consumers []string `json:"consumers"`
}

// ElasticsearchEndpoint represents the Elasticsearch endpoint to output logs.
type ElasticsearchEndpoint struct {
	// Hosts represents the array of Elasticsearch hosts.
	Hosts []string `json:"hosts"`
}

// LogEndpointStatus describes status of a log endpoint instance.
type LogEndpointStatus struct {
}
