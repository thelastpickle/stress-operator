# TLPStress

The tlp-stress operator manages TLPStress instances. There are a few top-level fields that divide the spec into logical sections. 

## StressConfig
StressConfig contains properties that map to tlp-stress command line options.

The following properties are supported in `spec.stressConfig`:

**workload**

*description:* The tlp-stress profile to run. Workloads includes:

* KeyValue
* BasicTimeSeries
* CountersWide
* LWT
* Locking
* Maps
* MaterializedViews
* RandomPartitionAccess
* UdtTimeSeries
* AllowFiltering

*type:* string

*required:* no

*default:* KeyValue

***

**consistencyLevel**

*description:* The consistency level for reads/writes.

*type:* string

*required:* no

*default:* LOCAL_ONE

***

**partitions**

*description:* The max value of the integer component of first partition key. See [human readable strings](#human-readable-strings) for acceptable values.

*type:* string

*required:* no

*default:* 0

***

**dataCenter**

*description:* The data center to which requests should be set.

*type:* string

*required:* no

*default:* N/A

***

**duration**

*description:* The duration of the stress test. See [human readable strings](#human-readable-strings) for acceptable values.

*type:* string

*required:* no

*default:* 0

***

**dropKeyspace**

*description:* Drop the keyspace before starting.

*ype:* boolean

*required:* no

*default:* false

***

**iterations**

*description:* Number of operations to perform. See [human readable strings](#human-readable-strings) for acceptable values.

*type:* string

*required:* no

*default:* 0

***

**readRate**

**populate**

**concurrency**

**partitionGenerator**

**replication** 

### Human Readable Strings
tlp-stress allows values for several parameters to be expressed as *human readable strings*. Instead of writing one billion as `1000000000`, you can instead write `1b`.

Human readable strings are validated using this regex:

```
^(\d+)([BbMmKk]?$)
```

The following fields support human readable strings:

* partitions
* iterations
* populate
* duration

Note that `duration` is a bit different than the other fields. It is not governed by the above regex and supports values like:

* 1h --> 1 hour
* 1m --> 1 minute
* 1d --> 1 day
* 3h 30m --> 3 hours, 30 minuntes
* 1d 6h 20m --> 1 day, 6 hours, 20 minutes 


## JobConfig

## CassandraConfig

## Status