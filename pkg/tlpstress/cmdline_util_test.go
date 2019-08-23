package tlpstress

import (
	"github.com/jsanda/tlp-stress-operator/pkg/apis/thelastpickle/v1alpha1"
	"reflect"
	"testing"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestCreateCommandLineArgs(t *testing.T) {
	type args struct {
		stressCfg    *v1alpha1.TLPStressConfig
		cassandraCfg *v1alpha1.CassandraConfig
		namespace    string
	}
	tests := []struct {
		name string
		args args
		want *CommandLineArgs
	}{
		{
			name: "AllOptionsExceptReplicationWithCassandraService",
			args: args{
				stressCfg: &v1alpha1.TLPStressConfig{
					Workload: v1alpha1.BasicTimeSeriesWorkload,
					ConsistencyLevel: v1alpha1.CL_LOCAL_QUORUM,
					Partitions: stringRef("3m"),
					Duration: "2h",
					DropKeyspace: true,
					Iterations: stringRef("10b"),
					DataCenter: "dc1",
					ReadRate: "0.25",
					Populate: stringRef("10m"),
					Concurrency: int32Ref(250),
					PartitionGenerator: "random",
				},
				cassandraCfg: &v1alpha1.CassandraConfig{
					CassandraService: "cassandra-test",
				},
				namespace: "default",
			},
			want: &CommandLineArgs{
				args: []string{
					"run", string(v1alpha1.BasicTimeSeriesWorkload),
					"--cl", string(v1alpha1.CL_LOCAL_QUORUM),
					"--partitions", "3m",
					"--duration", "2h",
					"--drop", "",
					"--iterations", "10b",
					"--readrate", "0.25",
					"--populate", "10m",
					"--concurrency", "250",
					"--partitiongenerator", "random",
					"--dc", "dc1",
					"--host", "cassandra-test",
				},
			},
		},
		{
			name: "SimpleStrategyReplicationWithCassandraService",
			args: args{
				stressCfg: &v1alpha1.TLPStressConfig{
					Workload: v1alpha1.KeyValueWorkload,
					Replication: v1alpha1.ReplicationConfig{
						SimpleStrategy: int32Ref(3),
					},
				},
				cassandraCfg: &v1alpha1.CassandraConfig{
					CassandraService: "cassandra-test",
				},
				namespace: "default",
			},
			want: &CommandLineArgs{
				args: []string{
					"run", string(v1alpha1.KeyValueWorkload),
					"--replication", "{'class': 'SimpleStrategy', 'replication_factor': 3}",
					"--host", "cassandra-test",
				},
			},
		},
		{
			name: "NetworkTopologyStrategyReplicationWithCassandraService",
			args: args{
				stressCfg: &v1alpha1.TLPStressConfig{
					Workload: v1alpha1.KeyValueWorkload,
					Replication: v1alpha1.ReplicationConfig{
						NetworkTopologyStrategy: &map[string]int32 {
							"dc1": 2,
						},
					},
				},
				cassandraCfg: &v1alpha1.CassandraConfig{
					CassandraService: "cassandra-test",
				},
				namespace: "default",
			},
			want: &CommandLineArgs{
				args: []string{
					"run", string(v1alpha1.KeyValueWorkload),
					"--replication", "{'class': 'NetworkTopologyStrategy', 'dc1': 2}",
					"--host", "cassandra-test",
				},
			},
		},
		{
			name: "WorkloadWithCassandraClusterWithoutNamespace",
			args: args{
				stressCfg: &v1alpha1.TLPStressConfig{
					Workload: v1alpha1.MapsWorkload,
				},
				cassandraCfg: &v1alpha1.CassandraConfig{
					CassandraCluster: &v1alpha1.CassandraCluster{
						Name: "cassandra-test",
					},
				},
			},
			want: &CommandLineArgs{
				args: []string{
					"run", string(v1alpha1.MapsWorkload),
					"--host", "cassandra-test",
				},
			},
		},
		{
			name: "WorkloadWithCassandraServiceWithNamespace",
			args: args{
				stressCfg: &v1alpha1.TLPStressConfig{
					Workload: v1alpha1.CountersWideWorkload,
				},
				cassandraCfg: &v1alpha1.CassandraConfig{
					CassandraCluster: &v1alpha1.CassandraCluster{
						Name: "cassandra-test",
						Namespace: "dev",
					},
				},
			},
			want: &CommandLineArgs{
				args: []string{
					"run", string(v1alpha1.CountersWideWorkload),
					"--host", "cassandra-test.dev.svc.cluster.local",
				},
			},
		},
		{
			name: "WorkloadWithCassandraClusterTemplateWithoutNamespace",
			args: args{
				stressCfg: &v1alpha1.TLPStressConfig{
					Workload: v1alpha1.LWTWorkload,
				},
				cassandraCfg: &v1alpha1.CassandraConfig{
					CassandraClusterTemplate: &v1alpha1.CassandraClusterTemplate{
						ObjectMeta: metav1.ObjectMeta{
							Name: "cassandra-test",
						},
					},
				},
			},
			want: &CommandLineArgs{
				args: []string{
					"run", string(v1alpha1.LWTWorkload),
					"--host", "cassandra-test",
				},
			},
		},
		{
			name: "WorkloadWithCassandraClusterTemplateWithNamespace",
			args: args{
				stressCfg: &v1alpha1.TLPStressConfig{
					Workload: v1alpha1.UdtTimeSeriesWorkload,
				},
				cassandraCfg: &v1alpha1.CassandraConfig{
					CassandraClusterTemplate: &v1alpha1.CassandraClusterTemplate{
						ObjectMeta: metav1.ObjectMeta{
							Name: "cassandra-test",
							Namespace: "dev",
						},
					},
				},
			},
			want: &CommandLineArgs{
				args: []string{
					"run", string(v1alpha1.UdtTimeSeriesWorkload),
					"--host", "cassandra-test.dev.svc.cluster.local",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CreateCommandLineArgs(tt.args.stressCfg, tt.args.cassandraCfg, tt.args.namespace); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CreateCommandLineArgs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func stringRef(s string) *string {
	return &s
}

func int32Ref(n int32) *int32 {
	return &n
}