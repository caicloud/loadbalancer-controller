package v1alpha1

// AlertingRulePhase is the phase of an alerting rule (of any kind)
type AlertingRulePhase string

const (
	// An alerting rule is ok when it is working and not firing.
	AlertingRulePhaseOk AlertingRulePhase = "OK"

	// An alerting rule is firing when its firing condition has been met.
	AlertingRulePhaseFiring AlertingRulePhase = "Firing"

	// An alerting rule lost its target when its ObjectReference cannot match any resources.
	AlertingRulePhaseTargetLost AlertingRulePhase = "TargetLost"

	// An alerting rule can be marked as disabled to disable all its associated functionality.
	AlertingRulePhaseDisabled AlertingRulePhase = "Disabled"
)

// Comparator is the comparator used to compare the comparable and threshold of an alerting rule.
type Comparator string

const (
	ComparatorGreater      Comparator = ">"
	ComparatorGreaterEqual Comparator = ">="
	ComparatorLesser       Comparator = "<"
	ComparatorLesserEqual  Comparator = "<="
	ComparatorEqual        Comparator = "=="
	ComparatorNotEqual     Comparator = "!="
)

// AlertingRuleKind denotes the kind of an alerting rule.
type AlertingRuleKind string

const (
	LogAlertingRule    AlertingRuleKind = "Log"
	MetricAlertingRule AlertingRuleKind = "Metric"
)

// ObjectKind describes the kind of resource that an alerting rule or sub rule is targeting.
type ObjectKind string

const (
	ObjectKindCluster   ObjectKind = "Cluster"
	ObjectKindNode      ObjectKind = "Node"
	ObjectKindNamespace ObjectKind = "Namespace"
	ObjectKindRelease   ObjectKind = "Release"
	ObjectKindPod       ObjectKind = "Pod"
	ObjectKindContainer ObjectKind = "Container"
)

// ObjectReference serves as a filter for resources; an alerting rule would apply itself on
// the resources that the ObjectReference matches. An ObjectReference is considered valid if
// and only if the combination of the non-empty fields describe the targeted kind exactly.
// For example, for `Container` kind, `Cluster`, `Namespace`, `Pod`, and `Container` should
// contain the appropriate names, and other fields should be empty string.
type ObjectReference struct {
	Kind      ObjectKind `json:"kind"`
	Cluster   string     `json:"cluster,omitempty"`
	Node      string     `json:"node,omitempty"`
	Namespace string     `json:"namespace,omitempty"`
	Release   string     `json:"release,omitempty"`
	Pod       string     `json:"pod,omitempty"`
	Container string     `json:"container,omitempty"`
}
