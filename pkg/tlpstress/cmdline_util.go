package tlpstress

import (
	"fmt"
	"github.com/jsanda/tlp-stress-operator/pkg/apis/thelastpickle/v1alpha1"
	"strconv"
	"strings"
)

type CommandLineArgs struct {
	args map[string]string

	argsSlice *[]string
}

func (c *CommandLineArgs) GetArgs() *[]string {
	empty := make([]string, 0)
	return &empty
}

func (c *CommandLineArgs) String() string {
	return strings.Join(*c.argsSlice, " ")
}

// Generates the arguments that are passed to the tlp-stress executable
func CreateCommandLineArgs(cfg *v1alpha1.TLPStressConfig, cassandraCfg *v1alpha1.CassandraConfig, namespace string) *CommandLineArgs {
	args:= make(map[string]string)

	args["run"] = string(cfg.Workload)

	if len(cfg.ConsistencyLevel) > 0 {
		args["--cl"] = string(cfg.ConsistencyLevel)
	}

	if cfg.Partitions != nil {
		args["--partitions"] = *cfg.Partitions
	}

	if len(cfg.Duration) > 0 {
		args["--duration"] = cfg.Duration
	}

	if cfg.DropKeyspace {
		args["--drop"] = ""
	}

	if cfg.Iterations != nil {
		args["--iterations"] = *cfg.Iterations
	}

	if len(cfg.ReadRate) > 0 {
		args["--readrate"] = cfg.ReadRate
	}

	if cfg.Populate != nil {
		args["--populate"] = *cfg.Populate
	}

	if cfg.Concurrency != nil && *cfg.Concurrency != 100 {
		args["--concurrency"] = strconv.FormatInt(int64(*cfg.Concurrency), 10)
	}

	if len(cfg.PartitionGenerator) > 0 {
		args["--partitiongenerator"] = cfg.PartitionGenerator
	}

	if len(cfg.DataCenter) > 0 {
		args["--dc"] = cfg.DataCenter
	}

	// TODO Need to make sure only one replication strategy is specified
	if cfg.Replication.SimpleStrategy != nil {
		replicationFactor := strconv.FormatInt(int64(*cfg.Replication.SimpleStrategy), 10)
		replication := fmt.Sprintf(`{'class': 'SimpleStrategy', 'replication_factor': %s}`, replicationFactor)
		args["--replication"] = replication
	} else if cfg.Replication.NetworkTopologyStrategy != nil {
		var sb strings.Builder
		dcs := make([]string, 0)
		for k, v := range *cfg.Replication.NetworkTopologyStrategy {
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