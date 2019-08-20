package v1alpha1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	batchv1 "k8s.io/api/batch/v1"
)

type Workload string

const (
	KeyValueWorkload             Workload = "KeyValue"
	BasicTimeSeriesWorkload      Workload = "BasicTimeSeries"
	CountersWideWorkload         Workload = "CountersWide"
	LWTWorkload                  Workload = "LWT"
	LockingWorkload              Workload = "Locking"
	MapsWorkload                 Workload = "Maps"
	MaterializedViewsWorkload    Workload = "MaterializedViews"
	RandomParitionAccessWorkload Workload = "RandomPartitionAccess"
	UdtTimeSeriesWorkload        Workload = "UdtTimeSeries"
)

type ConsistencyLevel string

const (
	CL_ANY          ConsistencyLevel = "ANY"
	CL_ONE          ConsistencyLevel = "ONE"
	CL_TWO          ConsistencyLevel = "TWO"
	CL_THREE        ConsistencyLevel = "THREE"
	CL_QUORUM       ConsistencyLevel = "QUORUM"
	CL_ALL          ConsistencyLevel = "ALL"
	CL_LOCAL_QUORUM ConsistencyLevel = "LOCAL_QUORUM"
	CL_EACH_QUORUM  ConsistencyLevel = "EACH_QUORUM"
	CL_LOCAL_ONE    ConsistencyLevel = "LOCAL_ONE"
	CL_SERIAL       ConsistencyLevel = "SERIAL"
	CL_LOCAL_SERIAL ConsistencyLevel = "LOCAL_SERIAL"
)

// Describes the data the the job should have
type JobConfig struct {
	// Specifies the number of retries before marking this job failed.
	// Defaults to 6
	// +optional
	BackoffLimit *int32 `json:"backoffLimit,omitempty" protobuf:"varint,7,opt,name=backoffLimit"`

	// Specifies the maximum desired number of pods the job should
	// run at any given time. The actual number of pods running in steady state will
	// be less than this number when ((.spec.completions - .status.successful) < .spec.parallelism),
	// i.e. when the work left to do is less than max parallelism.
	// More info: https://kubernetes.io/docs/concepts/workloads/controllers/jobs-run-to-completion/
	// +optional
	Parallelism *int32 `json:"parallelism,omitempty" protobuf:"varint,1,opt,name=parallelism"`
}

type TLPStressConfig struct {
	// +kubebuilder:validation:Enum=KeyValue,BasisTimeSeries,CountersWide,LWT,Locking,Maps,MaterializedViews,RandomPartitionAccess,UdtTimeSeries
	Workload Workload `json:"workload,omitempty"`

	// +kubebuilder:validation:Enum=ANY,ONE,TWO,THREE,QUORUM,ALL,LOCAL_QUORUM,EACH_QUORUM,SERIAL,LOCAL_SERIAL,LOCAL_ONE
	ConsistencyLevel ConsistencyLevel `json:"consistencyLevel,omitempty"`

	// +kubebuilder:validation:Pattern=^(\d+)([BbMmKk]?$)
	Partitions *string `json:"partitions,omitempty"`

	DataCenter string `json:"dataCenter,omitempty"`

	Duration string `json:"duration,omitempty"`

	DropKeyspace bool `json:"dropKeyspace,omitempty"`

	// +kubebuilder:validation:Pattern=^(\d+)([BbMmKk]?$)
	Iterations *string `json:"iterations,omitempty"`

	ReadRate string `json:"readRate,omitempty"`

	// +kubebuilder:validation:Pattern=^(\d+)([BbMmKk]?$)
	Populate *string `json:"populate,omitempty"`

	Concurrency *int32 `json:"concurrency,omitempty"`

	PartitionGenerator string `json:"partitionGenerator,omitempty"`

	Replication ReplicationConfig `json:"replication,omitempty"`
}

// TLPStressSpec defines the desired state of TLPStress
// +k8s:openapi-gen=true
type TLPStressSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html

	CassandraService string `json:"cassandraService,omitempty"`

	CassandraCluster *CassandraCluster `json:"cassandraCluster,omitempty"`

	Image string `json:"image"`

	ImagePullPolicy v1.PullPolicy `json:"imagePullPolicy"`

	StressConfig TLPStressConfig `json:"stressConfig,omitempty"`

	JobConfig JobConfig `json:"jobConfig,omitempty"`
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
	Namespace string `json:"namespace,omitempty"`
	Name      string `json:"name,omitempty"`
}

func init() {
	SchemeBuilder.Register(&TLPStress{}, &TLPStressList{})
}
