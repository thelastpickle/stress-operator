package v1alpha1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	batchv1 "k8s.io/api/batch/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// TLPStressSpec defines the desired state of TLPStress
// +k8s:openapi-gen=true
type TLPStressSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html

	// +kubebuilder:validation:Enum=KeyValue,BasisTimeSeries,CountersWide,LWT,Locking,Maps,MaterializedViews,RandomPartitionAccess,UdtTimeSeries
	Workload string `json:"workload"`

	// +kubebuilder:validation:Enum=ANY,ONE,TWO,THREE,QUORUM,ALL,LOCAL_QUORUM,EACH_QUORUM,SERIAL,LOCAL_SERIAL,LOCAL_ONE
	ConsistencyLevel string `json:"consistencyLevel,omitempty"`

	// +kubebuilder:validation:Minimum=1
	Partitions *int64 `json:"partitions,omitempty"`

	// +kubebuilder:validation:Pattern=(\d+)([BbMmKk])?
	Duration string `json:"duration,omitempty"`

	DropKeyspace *bool `json:"dropKeyspace,omitempty"`

	// +kubebuilder:validation:Minimum=0
	Iterations *int64 `json:"iterations,omitempty"`

	ReadRate  string `json:"readRate,omitempty"`

	// +kubebuilder:validation:Minimum=0
	Populate *int64 `json:"populate,omitempty"`

	// +kubebuilder:validation:Minimum=1
	Concurrency *int32 `json:"concurrency,omitempty"`

	PartitionGenerator string `json:"partitionGenerator,omitempty"`

	Replication ReplicationConfig `json:"replication,omitempty"`

	CassandraService string `json:"cassandraService"`

	CassandraCluster *CassandraCluster `json:"cassandraCluster,omitempty"`

	Image string `json:"image"`

	ImagePullPolicy v1.PullPolicy `json:"imagePullPolicy"`
}

type ReplicationConfig struct {
	SimpleStrategy *int32 `json:"simpleStrategy,omitempty"`

	NetworkTopologyStrategy *map[string]int32 `json:"networkTopologyStrategy,omitempty"`
}

// TLPStressStatus defines the observed state of TLPStress
// +k8s:openapi-gen=true
type TLPStressStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html

	JobStatus  *batchv1.JobStatus `json:"jobStatus,omitempty"`
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

type CassandraCluster struct {
	Namespace string
	Name      string
}

func init() {
	SchemeBuilder.Register(&TLPStress{}, &TLPStressList{})
}
