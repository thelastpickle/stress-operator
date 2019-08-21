package tlpstress

import (
	"fmt"
	"github.com/jsanda/tlp-stress-operator/pkg/apis/thelastpickle/v1alpha1"
	"strconv"
	"strings"
)

type CommandLineArgs struct {
	args *[]string
}

func (c *CommandLineArgs) GetArgs() *[]string {
	return c.args
}

func (c *CommandLineArgs) String() string {
	return strings.Join(*c.args, " ")
}

// Generates the arguments that are passed to the tlp-stress executable
func CreateCommandLineArgs(instance *v1alpha1.TLPStress, namespace string) *CommandLineArgs {
	cfg := instance.Spec.StressConfig
	args:= make([]string, 0)

	args = append(args,"run")
	args = append(args, string(cfg.Workload))

	if len(cfg.ConsistencyLevel) > 0 {
		args = append(args, "--cl")
		args = append(args, string(cfg.ConsistencyLevel))
	}

	if cfg.Partitions != nil {
		args = append(args, "-p")
		args = append(args, *cfg.Partitions)
	}

	if len(cfg.Duration) > 0 {
		args = append(args, "-d")
		args = append(args, cfg.Duration)
	}

	if cfg.DropKeyspace {
		args = append(args, "--drop")
	}

	if cfg.Iterations != nil {
		args = append(args, "-n")
		args = append(args, *cfg.Iterations)
	}

	if len(cfg.ReadRate) > 0 {
		args = append(args, "-r")
		args = append(args, cfg.ReadRate)
	}

	if cfg.Populate != nil {
		args = append(args, "--populate")
		args = append(args, *cfg.Populate)
	}

	if cfg.Concurrency != nil && *cfg.Concurrency != 100 {
		args = append(args, "-c")
		args = append(args, strconv.FormatInt(int64(*cfg.Concurrency), 10))
	}

	if len(cfg.PartitionGenerator) > 0 {
		args = append(args, "--pg")
		args = append(args, cfg.PartitionGenerator)
	}

	if len(cfg.DataCenter) > 0 {
		args = append(args, "--dc")
		args = append(args, cfg.DataCenter)
	}

	// TODO Need to make sure only one replication strategy is specified
	if cfg.Replication.SimpleStrategy != nil {
		replicationFactor := strconv.FormatInt(int64(*cfg.Replication.SimpleStrategy), 10)
		replication := fmt.Sprintf(`{'class': 'SimpleStrategy', 'replication_factor': %s}`, replicationFactor)
		args = append(args, "--replication")
		args = append(args, replication)
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
		args = append(args, "--replication")
		args = append(args, replication)
	}

	// TODO add validation check that either CassandraService or CassandraCluster is defined in the spec
	svc := ""
	if instance.Spec.CassandraCluster != nil {
		// The headless service for a CassandraCluster has the same name as the cluster
		if instance.Spec.CassandraCluster.Namespace == "" || instance.Spec.CassandraCluster.Namespace == namespace {
			svc = instance.Spec.CassandraCluster.Name
		} else {
			// CassandraCluster service is in a different namespace
			svc = fmt.Sprintf("%s.%s.svc.cluster.local", instance.Spec.CassandraCluster.Name,
				instance.Spec.CassandraCluster.Name)
		}
	} else {
		svc = instance.Spec.CassandraService
	}
	args = append(args, "--host")
	args = append(args, svc)

	return &CommandLineArgs{args: &args}
}