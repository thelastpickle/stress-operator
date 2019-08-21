package tlpstress

import (
	"fmt"
	"github.com/jsanda/tlp-stress-operator/pkg/apis/thelastpickle/v1alpha1"
	"strconv"
	"strings"
)

type CommandLineArgs struct {
	args map[string]string
}

func (c *CommandLineArgs) GetArgs() *[]string {
	empty := make([]string, 0)
	return &empty
}

func (c *CommandLineArgs) String() string {
	//return strings.Join(*c.argsSlice, " ")
	return fmt.Sprint(c.args)
}

// Generates the arguments that are passed to the tlp-stress executable
func CreateCommandLineArgs(stressCfg *v1alpha1.TLPStressConfig, cassandraCfg *v1alpha1.CassandraConfig, namespace string) *CommandLineArgs {
	args:= make(map[string]string)

	args["run"] = string(stressCfg.Workload)

	if len(stressCfg.ConsistencyLevel) > 0 {
		args["--cl"] = string(stressCfg.ConsistencyLevel)
	}

	if stressCfg.Partitions != nil {
		args["--partitions"] = *stressCfg.Partitions
	}

	if len(stressCfg.Duration) > 0 {
		args["--duration"] = stressCfg.Duration
	}

	if stressCfg.DropKeyspace {
		args["--drop"] = ""
	}

	if stressCfg.Iterations != nil {
		args["--iterations"] = *stressCfg.Iterations
	}

	if len(stressCfg.ReadRate) > 0 {
		args["--readrate"] = stressCfg.ReadRate
	}

	if stressCfg.Populate != nil {
		args["--populate"] = *stressCfg.Populate
	}

	if stressCfg.Concurrency != nil && *stressCfg.Concurrency != 100 {
		args["--concurrency"] = strconv.FormatInt(int64(*stressCfg.Concurrency), 10)
	}

	if len(stressCfg.PartitionGenerator) > 0 {
		args["--partitiongenerator"] = stressCfg.PartitionGenerator
	}

	if len(stressCfg.DataCenter) > 0 {
		args["--dc"] = stressCfg.DataCenter
	}

	// TODO Need to make sure only one replication strategy is specified
	if stressCfg.Replication.SimpleStrategy != nil {
		replicationFactor := strconv.FormatInt(int64(*stressCfg.Replication.SimpleStrategy), 10)
		replication := fmt.Sprintf(`{'class': 'SimpleStrategy', 'replication_factor': %s}`, replicationFactor)
		args["--replication"] = replication
	} else if stressCfg.Replication.NetworkTopologyStrategy != nil {
		var sb strings.Builder
		dcs := make([]string, 0)
		for k, v := range *stressCfg.Replication.NetworkTopologyStrategy {
			sb.WriteString("'")
			sb.WriteString(k)
			sb.WriteString("': ")
			sb.WriteString(strconv.FormatInt(int64(v), 10))
			dcs = append(dcs, sb.String())
			sb.Reset()
		}
		replication := fmt.Sprintf("{'class': 'NetworkTopologyStrategy', %s}", strings.Join(dcs, ", "))
		args["--replication"] = replication
	}

	// TODO add validation check that either CassandraService or CassandraCluster is defined in the spec
	svc := ""
	if cassandraCfg.CassandraCluster != nil {
		// The headless service for a CassandraCluster has the same name as the cluster
		if cassandraCfg.CassandraCluster.Namespace == "" || cassandraCfg.CassandraCluster.Namespace == namespace {
			svc = cassandraCfg.CassandraCluster.Name
		} else {
			// CassandraCluster service is in a different namespace
			svc = fmt.Sprintf("%s.%s.svc.cluster.local", cassandraCfg.CassandraCluster.Name,
				cassandraCfg.CassandraCluster.Name)
		}
	} else {
		svc = cassandraCfg.CassandraService
	}
	args["--host"] = svc

	return &CommandLineArgs{args: args}
}