# Storage

On-disk files in a container are ephemeral, which presents some problems for non-trivial applications when running in containers. First, when a container crashes, kubelet will restart it but the files will be lost - the container starts with a clean state. Second, when running containers together in a Pod it is often necessary to share files between those containers. The Kubernetes Volume abstraction solves both of these problems.

## Volumes

A Kubernetes volume has an explicit lifetime - the same as the Pod that encloses it. Consequently, a volume outlives any Containers that run within the Pod, and data is preserved across Container restarts. When a Pod ceases to exist, the volume will cease to exist, too.

## Persistent Volumes

A PersistentVolume (PV) is a piece of storage in the cluster that has been provisioned by an administrator or dynamically provisioned using Storage Classes. It is a resource in the cluster just like a node is a cluster resource. PVs are volume plugins like Volumes, but have a lifecycle independent of any individual Pod that uses the PV.

## Persistent Volume Claims

A PersistentVolumeClaim (PVC) is a request for storage by a user. It is similar to a Pod. Pods consume node resources and PVCs consume PV resources. Pods can request specific levels of resources (CPU and Memory). Claims can request specific size and access modes (e.g., they can be mounted ReadWriteOnce, ReadOnlyMany or ReadWriteMany).
