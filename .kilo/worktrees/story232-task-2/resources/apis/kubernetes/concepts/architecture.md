# Kubernetes Cluster Architecture

Kubernetes follows a client-server architecture. It's possible to have a multi-master setup (for high availability), but by default there is a single master node which acts as a controlling node and point of contact. The master node manages the cluster and the worker nodes are the ones which run the applications.

## Components

### Control Plane Components

The control plane's components make global decisions about the cluster (for example, scheduling), as well as detecting and responding to cluster events (for example, starting up a new pod when a deployment's `replicas` field is unsatisfied).

- **kube-apiserver**: The API server is a component of the Kubernetes control plane that exposes the Kubernetes API. The API server is the front end for the Kubernetes control plane.
- **etcd**: Consistent and highly-available key value store used as Kubernetes' backing store for all cluster data.
- **kube-scheduler**: Control plane component that watches for newly created Pods with no assigned node, and selects a node for them to run on.
- **kube-controller-manager**: Control plane component that runs controller processes. Logically, each controller is a separate process, but to reduce complexity, they are all compiled into a single binary and run in a single process.
- **cloud-controller-manager**: A Kubernetes control plane component that embeds cloud-specific control logic. The cloud controller manager lets you link your cluster into your cloud provider's API, and separates out the components that interact with that cloud platform from components that just interact with your cluster.

### Node Components

Node components run on every node, maintaining running pods and providing the Kubernetes runtime environment.

- **kubelet**: An agent that runs on each node in the cluster. It makes sure that containers are running in a Pod.
- **kube-proxy**: kube-proxy is a network proxy that runs on each node in your cluster, implementing part of the Kubernetes Service concept.
- **Container Runtime**: The container runtime is the software that is responsible for running containers.
