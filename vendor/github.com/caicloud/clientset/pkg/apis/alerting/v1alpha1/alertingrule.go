package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AlertingRule describes an alerting rule. It holds the template for a list of AlertingSubRules and
// data required to manage and utilize them.
// All AlertingRule are located in the control cluster.
type AlertingRule struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              AlertingRuleSpec   `json:"spec"`
	Status            AlertingRuleStatus `json:"status"`
}

// AlertingRuleSpec is the spec for a AlertingRule.
type AlertingRuleSpec struct {
	// Template is the template from which the AlertingSubRules should be created.
	Template *SubRuleTemplate `json:"template"`

	// NotifyTypes is a list of notifying methods to be used when sending notifications. (NotifyType is
	// defined in NotifyAdmin and managed there)
	NotifyTypes []string `json:"notifyTypes,omitempty"`

	// ContactGroups is a list of ContactGroup IDs where the alert notifications should go. (ContactGroup is
	// defined in NotifyAdmin and managed there)
	ContactGroups []int `json:"contactGroups,omitempty"`
}

// AlertingRuleStatus describes the status of a AlertingRule.
type AlertingRuleStatus struct {
	Phase AlertingRulePhase `json:"phase"`

	// PhaseTimestamp is the time when the rule changed to its current phase.
	PhaseTimestamp *metav1.Time `json:"phaseTimestamp"`

	// FiringSubRules is a list of firing sub rules that should be handled.
	FiringSubRules []AlertingSubRule `json:"firingSubRules,omitempty"`
}

// SubRuleTemplate is the template from which the AlertingSubRules should be created. An example of
// such template can be:
//{
//	"target": {
//		"kind": "Namespace",
//		"cluster": "compass-stack",
//		"namespace": "default"
//	},
//  "kind": "Metric",
//	"rules": [
//		{
//			"comparator": ">",
//			"threshold": "1073741824",
//			"comparable": "sum(app:container_memory_rss{namespace=\"default\"}) by (pod_name)",
//			"range": 300,
//			"level": 0,
//			"notify": false
//		},
//		{
//			"comparator": ">",
//			"threshold": "2147483648",
//			"comparable": "sum(app:container_memory_rss{namespace=\"default\"}) by (pod_name)",
//			"range": 300,
//			"level": 1,
//			"notify": true
//		}
//	]
//}
type SubRuleTemplate struct {
	// All sub rules created from this template share this `Target` field
	Target *ObjectReference `json:"target"`

	// All sub rules created from this template share this `Kind` field
	Kind AlertingRuleKind `json:"kind"`

	// CheckInterval is the interval to recheck the firing condition.
	CheckInterval string `json:"checkInterval,omitempty"`

	// A list of specs for the AlertingSubRules under the AlertingRule. The `Target` and `Kind` field
	// of these specs should be nil (if they are not, they should be ignored).
	Rules []AlertingSubRuleSpec `json:"rules"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AlertingRuleList describes a list of AlertingRuleList.
type AlertingRuleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []AlertingRule `json:"items"`
}
