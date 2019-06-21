package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AlertingSubRule describes a sub rule of an AlertingRule. A sub rules represents one level of
// an alerting rule. Its `metadata.ownerReference` points to the AleringRule that created it.
// All AlertingSubRule are located in the control cluster.
type AlertingSubRule struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              AlertingSubRuleSpec   `json:"spec"`
	Status            AlertingSubRuleStatus `json:"status,omitempty"`
}

// AlertingSubRuleSpec is the spec for an AlertingSubRule
type AlertingSubRuleSpec struct {
	// Condition describes condition for the rule to go into Firing phase.
	Condition *FireCondition `json:"condition"`

	// Target describes the target of the rule.
	Target *ObjectReference `json:"target,omitempty"`

	// CheckInterval is the interval to recheck the firing condition.
	CheckInterval string `json:"checkInterval,omitempty"`

	// Kind denotes whether this is a log alerting rule or a metric one.
	Kind AlertingRuleKind `json:"kind"`

	// Level is the severity level of this sub rule. In principle, the lower the number, the less
	// severe it is; definitions of the levels are subjected to change and thus defined where the
	// rule is used.
	Level int `json:"level"`

	// Notify denotes whether or not this sub rules should produce notifications.
	Notify bool `json:"notify"`
}

// AlertingSubRuleStatus describes the status of an AlertingSubRule
type AlertingSubRuleStatus struct {
	Phase AlertingRulePhase `json:"phase"`

	// PhaseTimestamp is the time when the sub rule changed to its current phase.
	PhaseTimestamp *metav1.Time `json:"phaseTimestamp"`

	// AlertFingerprint is a unique value that is hashed from other fields of the alerting sub rule
	// and refreshed it enters the 'Ok' phase. New alerts generated from a alerting sub rule would be
	// branded with this fingerprint. It is used to identify duplicated and outdated alerts.
	AlertFingerprint string `json:"alertFingerprint"`
}

// FireCondition describes the requirements for an alerting rule to go into Firing phase.
type FireCondition struct {
	Comparator Comparator `json:"comparator"`
	Threshold  float64    `json:"threshold"`

	// For LogAlertRules, Comparable is the keywords to look for
	// For MetricAlertRules, Comparable is a Prometheus query
	Comparable string `json:"comparable"`

	// For LogAlertRules, Range is the time range in which to search the keywords.
	// For MetricAlertRules, a Prometheus alerting rule must stay pending for at least this amount of time
	// before it actually fires.
	Range string `json:"range"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AlertingSubRuleList describes a list of AlertingSubRules
type AlertingSubRuleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []AlertingSubRule `json:"items"`
}
