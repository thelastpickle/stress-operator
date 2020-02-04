# CassKop Integration
The stress-operator provides the ability to provision a Cassandra cluster using the [CassKop operator](./casskop.md). CassKop support is built directly into the stress-operator. The CassKop operator however is an optional dependency.

`spec.cassandraConfig` has a couple property that deal directly with CassKop - `cassandraCluster` and `cassandraClusterTemplate`.

## CassandraCluster
`spec.cassandraConfig.cassandraCluster` specifies the name of a `CassandraCluster` object that is managed by CassKop. Because the stress-operator has built-in support for CassKop, it can use the cluster name to discover endpoints of the Cassandra pods.

`cassandraCluster` is comprised of the following properties:

**namespace**

*description:* The namespace in which the `CassandraCluster` is running. Defaults to the namespace in which the `Stress` instance runs.

*type:* string

*required:* no

*default:* N/A

***

**name**

*description:* The name of the `CassandraCluster`.

*required:* no

*default:* N/A

## cassandraClusterTemplate
`spec.cassandraConfig.cassandraClusterTemplate` describes a `CassandraCluster` that will be created prior to creating and running the `Stress` instance.

The stress-operator will not deploy the underlying job for the `Stress` instance until the `CassandraCluster` is fully initialized and ready.

Check out the [CassKop docs](https://github.com/Orange-OpenSource/cassandra-k8s-operator/blob/master/documentation/description.md) for a thorough description of how to configure a `CassandraCluster` object.

The following introductory blog posts also provide some good background:

* [Deploying Cassandra in Kubernetes with CassKop](https://medium.com/@john.sanda/deploying-cassandra-in-kubernetes-with-casskop-1f721586825b)
* [Create Affinity between Cassandra and Kubernetes](https://medium.com/@john.sanda/create-affinity-between-cassandra-and-kubernetes-5b2f72d23293)
