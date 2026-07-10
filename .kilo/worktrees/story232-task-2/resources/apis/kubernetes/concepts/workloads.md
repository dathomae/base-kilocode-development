# Workloads

A workload is an application running on Kubernetes. Whether your workload is a single component or several that work together, on Kubernetes you run it inside a set of Pods. In Kubernetes, a Pod represents a set of running containers on your cluster.

## Pods

A Pod is the smallest deployable unit of computing that you can create and manage in Kubernetes. A Pod (as in a pod of whales or pea pod) is a group of one or more containers, with shared storage and network resources, and a specification for how to run the containers.

## Workload Resources

Kubernetes provides several built-in workload resources:

- **Deployment**: A Deployment provides declarative updates for Pods and ReplicaSets. You describe a desired state in a Deployment, and the Deployment Controller changes the actual state to the desired state at a controlled rate.
- **ReplicaSet**: A ReplicaSet's purpose is to maintain a stable set of replica Pods running at any given time. As such, it is often used to guarantee the availability of a specified number of identical Pods.
- **StatefulSet**: StatefulSet is the workload API object used to manage stateful applications. Manages the deployment and scaling of a set of Pods, and provides guarantees about the ordering and uniqueness of these Pods.
- **DaemonSet**: A DaemonSet ensures that all (or some) Nodes run a copy of a Pod. As nodes are added to the cluster, Pods are added to them. As nodes are removed from the cluster, those Pods are garbage collected.
- **Job**: A Job creates one or more Pods and will continue to retry execution of the Pods until a specified number of them successfully terminate.
- **CronJob**: A CronJob creates Jobs on a repeating schedule.
