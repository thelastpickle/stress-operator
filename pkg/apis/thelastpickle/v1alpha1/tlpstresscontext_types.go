package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// TLPStressContextSpec defines the desired state of TLPStressContext
// +k8s:openapi-gen=true
type TLPStressContextSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html

	InstallPrometheus bool `json:"installPrometheus,omitempty"`

	InstallGrafana bool `json:"installGrafana,omitempty"`
}

// TLPStressContextStatus defines the observed state of TLPStressContext
// +k8s:openapi-gen=true
type TLPStressContextStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// TLPStressContext is the Schema for the tlpstresscontexts API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
type TLPStressContext struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TLPStressContextSpec   `json:"spec,omitempty"`
	Status TLPStressContextStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// TLPStressContextList contains a list of TLPStressContext
type TLPStressContextList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TLPStressContext `json:"items"`
}

func init() {
	SchemeBuilder.Register(&TLPStressContext{}, &TLPStressContextList{})
}
