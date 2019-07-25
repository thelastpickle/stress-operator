package v1alpha1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// TLPStressSpec defines the desired state of TLPStress
// +k8s:openapi-gen=true
type TLPStressSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html

	Workload string `json:"workload"`

	CassandraService string `json:"cassandraService"`

	Image string `json:"image"`

	ImagePullPolicy v1.PullPolicy `json:"imagePullPolicy"`
}

// TLPStressStatus defines the observed state of TLPStress
// +k8s:openapi-gen=true
type TLPStressStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html

	Status string `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// TLPStress is the Schema for the tlpstresses API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
type TLPStress struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TLPStressSpec   `json:"spec,omitempty"`
	Status TLPStressStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// TLPStressList contains a list of TLPStress
type TLPStressList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TLPStress `json:"items"`
}

func init() {
	SchemeBuilder.Register(&TLPStress{}, &TLPStressList{})
}
