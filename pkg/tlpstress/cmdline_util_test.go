package tlpstress

import (
	"github.com/jsanda/tlp-stress-operator/pkg/apis/thelastpickle/v1alpha1"
	"reflect"
	"testing"
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
				args: map[string]string{
					"run": string(v1alpha1.BasicTimeSeriesWorkload),
					"--host": "cassandra-test",
					"--cl": string(v1alpha1.CL_LOCAL_QUORUM),
					"--partitions": "3m",
					"--duration": "2h",
					"--drop": "",
					"--iterations": "10b",
					"--dc": "dc1",
					"--readrate": "0.25",
					"--populate": "10m",
					"--concurrency": "250",
					"--partitiongenerator": "random",

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