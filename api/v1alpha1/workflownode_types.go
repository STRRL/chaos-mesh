// Copyright 2021 Chaos Mesh Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	LabelControlledBy = "chaos-mesh.org/controlled-by"
	LabelWorkflow     = "chaos-mesh.org/workflow"
)

// +kubebuilder:object:root=true
// +kubebuilder:resource:shortName=wfn
// +kubebuilder:subresource:status
type WorkflowNode struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the behavior of a node of workflow
	Spec WorkflowNodeSpec `json:"spec"`

	// +optional
	// Most recently observed status of the workflow node
	Status WorkflowNodeStatus `json:"status"`
}

type WorkflowNodeSpec struct {
	TemplateName string       `json:"template_name"`
	WorkflowName string       `json:"workflow_name"`
	Type         TemplateType `json:"type"`
	StartTime    *metav1.Time `json:"start_time"`
	// +optional
	Deadline *metav1.Time `json:"deadline,omitempty"`
	// +optional
	Tasks []string `json:"tasks,omitempty"`
	// +optional
	ConditionalTasks []ConditionalTask `json:"conditionalTasks,omitempty"`
	// +optional
	*EmbedChaos `json:",inline,omitempty"`
}

type WorkflowNodeStatus struct {

	// ChaosResource refs to the real chaos CR object.
	// +optional
	ChaosResource *corev1.TypedLocalObjectReference `json:"chaos_resource,omitempty"`

	// ConditionalBranches records the evaluation result of each ConditionalTask
	// +optional
	ConditionalBranches []ConditionalBranchesStatus `json:"conditionalBranches,omitempty"`

	// ActiveChildren means the created children node
	// +optional
	ActiveChildren []corev1.LocalObjectReference `json:"active_children,omitempty"`

	// Children is necessary for representing the order when replicated child template references by parent template.
	// +optional
	FinishedChildren []corev1.LocalObjectReference `json:"finished_children,omitempty"`

	// Represents the latest available observations of a workflow node's current state.
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	Conditions []WorkflowNodeCondition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

type ConditionalTask struct {
	Task string `json:"task"`
	// +optional
	Expression string `json:"expression,omitempty"`
}

type ConditionalBranchesStatus struct {
	// +optional
	Branches []ConditionalBranch `json:"branches"`
	// +optional
	Context []string `json:"context"`
}

type ConditionalBranch struct {
	Task             string `json:"task"`
	EvaluationResult bool   `json:"run"`
}

type WorkflowNodeConditionType string

const (
	ConditionAccomplished   WorkflowNodeConditionType = "Accomplished"
	ConditionDeadlineExceed WorkflowNodeConditionType = "DeadlineExceed"
	ConditionChaosInjected  WorkflowNodeConditionType = "ChaosInjected"
)

type WorkflowNodeCondition struct {
	Type   WorkflowNodeConditionType `json:"type"`
	Status corev1.ConditionStatus    `json:"status"`
	Reason string                    `json:"reason"`
}

// +kubebuilder:object:root=true
type WorkflowNodeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []WorkflowNode `json:"items"`
}

func init() {
	SchemeBuilder.Register(&WorkflowNode{}, &WorkflowNodeList{})
}

// Reasons
const (
	EntryCreated          string = "EntryCreated"
	InvalidEntry          string = "InvalidEntry"
	NodeAccomplished      string = "NodeAccomplished"
	NodeDeadlineExceed    string = "NodeDeadlineExceed"
	NodeDeadlineNotExceed string = "NodeDeadlineNotExceed"
	NodeDeadlineOmitted   string = "NodeDeadlineOmitted"
	ChaosCRCreated        string = "ChaosCRCreated"
	ChaosCRCreateFailed   string = "ChaosCRCreateFailed"
)
