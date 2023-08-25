package v1alpha1

import (
	"github.com/hobbyfarm/hf-provisioner-shared/retries"
	"github.com/rancher/wrangler/pkg/condition"
	"github.com/rancher/wrangler/pkg/genericcondition"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	ConditionInstanceExists  = condition.Cond("EC2InstanceExists")
	ConditionInstanceRunning = condition.Cond("EC2InstanceRunning")
	ConditionInstanceUpdated = condition.Cond("EC2InstanceUpdated")

	ConditionKeyPairImported = condition.Cond("KeyExists")
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Instance struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,inline"`

	Spec   InstanceSpec   `json:"spec"`
	Status InstanceStatus `json:"status"`
}

// +k8s:deepcopy-gen=true

type InstanceSpec struct {
	Machine  string  `json:"machine"`
	Instance v1.JSON `json:"instance"`
}

// +k8s:deepcopy-gen=true

type InstanceStatus struct {
	Instance   v1.JSON                             `json:"instance"`
	Conditions []genericcondition.GenericCondition `json:"conditions"`
	Retries    []retries.GenericRetry              `json:"retries"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type InstanceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Instance `json:"items,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type KeyPair struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,inline"`

	Spec   KeyPairSpec   `json:"spec"`
	Status KeyPairStatus `json:"status"`
}

// +k8s:deepcopy-gen=true

type KeyPairSpec struct {
	Machine string `json:"machine"`
	Secret  string `json:"secret"`

	Key v1.JSON `json:"key"`
}

// +k8s:deepcopy-gen=true

type KeyPairStatus struct {
	Key        v1.JSON                             `json:"key"`
	Conditions []genericcondition.GenericCondition `json:"conditions"`
	Retries    []retries.GenericRetry              `json:"retries"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type KeyPairList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []KeyPair `json:"items,omitempty"`
}
