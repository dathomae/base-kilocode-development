# Kind (Kubernetes in Docker)

`kind` is a tool for running local Kubernetes clusters using Docker container "nodes". It was primarily designed for testing Kubernetes itself, but may be used for local development or CI.

## Prerequisites

- **Docker**: `kind` requires Docker to be installed and running.
  - [Install Docker on Ubuntu](https://docs.docker.com/engine/install/ubuntu/)
- **kubectl**: While not strictly required to run `kind`, it is necessary to interact with the cluster.
  - [Install kubectl](kubectl.md)

## Installation on Ubuntu

### Installing from Binary

1.  **Download the binary:**

    ```bash
    # For AMD64 / x86_64
    [ $(uname -m) = x86_64 ] && curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.31.0/kind-linux-amd64

    # For ARM64
    [ $(uname -m) = aarch64 ] && curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.31.0/kind-linux-arm64
    ```

2.  **Make the binary executable:**

    ```bash
    chmod +x ./kind
    ```

3.  **Move the binary to your PATH:**

    ```bash
    sudo mv ./kind /usr/local/bin/kind
    ```

4.  **Verify installation:**

    ```bash
    kind --version
    ```

## Basic Usage

### Creating a Cluster

To create a default cluster:

```bash
kind create cluster
```

This will bootstrap a Kubernetes cluster using a pre-built node image. By default, the cluster will be named `kind`.

To create a cluster with a specific name:

```bash
kind create cluster --name my-cluster
```

To create a cluster with a specific Kubernetes version:

```bash
kind create cluster --image kindest/node:v1.31.0
```

(Check [Releases](https://github.com/kubernetes-sigs/kind/releases) for available images)

### Interacting with the Cluster

`kind` automatically merges the cluster configuration into your `~/.kube/config` (or creates it if it doesn't exist). You can use `kubectl` to interact with the cluster.

```bash
kubectl cluster-info --context kind-kind
kubectl get nodes
kubectl get pods -A
```

If you created a named cluster (e.g., `my-cluster`), the context name will be `kind-my-cluster`.

### Loading Images

Since `kind` runs in Docker, it cannot see images on your host machine unless you load them. This avoids the need to push images to a registry.

```bash
kind load docker-image my-app:latest
```

If using a named cluster:

```bash
kind load docker-image my-app:latest --name my-cluster
```

### Deleting a Cluster

To delete the default cluster:

```bash
kind delete cluster
```

To delete a named cluster:

```bash
kind delete cluster --name my-cluster
```

## Multi-Node Clusters

You can configure `kind` to run a multi-node cluster (simulating multiple machines) using a YAML config file.

**Example `kind-config.yaml`:**

```yaml
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
  - role: control-plane
  - role: worker
  - role: worker
```

Create the cluster:

```bash
kind create cluster --config kind-config.yaml
```

## Ingress

To use Ingress (e.g., NGINX) in `kind`, you need to map ports 80 and 443 from the container to the host and label the node.

1.  **Create a config file `kind-ingress.yaml`:**

    ```yaml
    kind: Cluster
    apiVersion: kind.x-k8s.io/v1alpha4
    nodes:
      - role: control-plane
        kubeadmConfigPatches:
          - |
            kind: InitConfiguration
            nodeRegistration:
              kubeletExtraArgs:
                node-labels: "ingress-ready=true"
        extraPortMappings:
          - containerPort: 80
            hostPort: 80
            protocol: TCP
          - containerPort: 443
            hostPort: 443
            protocol: TCP
    ```

2.  **Create the cluster:**

    ```bash
    kind create cluster --config kind-ingress.yaml
    ```

3.  **Install NGINX Ingress Controller:**

    ```bash
    kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/kind/deploy.yaml
    ```

## LoadBalancer (MetalLB)

`kind` does not implement a LoadBalancer by default. You can use MetalLB to provide this.

1.  **Install MetalLB:**

    ```bash
    kubectl apply -f https://raw.githubusercontent.com/metallb/metallb/v0.13.7/config/manifests/metallb-native.yaml
    ```

2.  **Wait for pods to be ready:**

    ```bash
    kubectl wait --namespace metallb-system \
                  --for=condition=ready pod \
                  --selector=app=metallb \
                  --timeout=90s
    ```

3.  **Configure the IP Address Pool:**

    First, get the Docker network subnet:

    ```bash
    docker network inspect -f '{{.IPAM.Config}}' kind
    ```

    Create a `metallb-config.yaml` (adjust the addresses to be within the subnet output above, e.g., if subnet is `172.18.0.0/16`, use `172.18.255.200-172.18.255.250`):

    ```yaml
    apiVersion: metallb.io/v1beta1
    kind: IPAddressPool
    metadata:
      name: example
      namespace: metallb-system
    spec:
      addresses:
        - 172.18.255.200-172.18.255.250
    ---
    apiVersion: metallb.io/v1beta1
    kind: L2Advertisement
    metadata:
      name: empty
      namespace: metallb-system
    ```

    Apply it:

    ```bash
    kubectl apply -f metallb-config.yaml
    ```

## Local Registry

Using a local Docker registry allows you to push images locally and use them in `kind` without pushing to a remote registry like Docker Hub.

1.  **Start a local registry container:**

    ```bash
    docker run -d --restart=always -p "127.0.0.1:5001:5000" --name "kind-registry" registry:2
    ```

2.  **Create a cluster config `kind-registry.yaml`:**

    ```yaml
    kind: Cluster
    apiVersion: kind.x-k8s.io/v1alpha4
    containerdConfigPatches:
      - |-
        [plugins."io.containerd.grpc.v1.cri".registry]
          config_path = "/etc/containerd/certs.d"
    ```

3.  **Create the cluster and connect the registry:**

    ```bash
    kind create cluster --config kind-registry.yaml

    # Connect the registry to the cluster network if not already connected
    docker network connect "kind" "kind-registry" || true

    # Document the local registry
    kubectl apply -f https://raw.githubusercontent.com/kubernetes-sigs/kind/main/site/static/examples/kind-with-registry.yaml
    ```

## Known Issues on Linux/Ubuntu

- **File Watchers (inotify):**
  Kubernetes and Kind can consume a lot of inotify watches. If you see errors related to "too many open files", you may need to increase the limit.

  ```bash
  # Check current limit
  sysctl fs.inotify.max_user_watches

  # Increase limit (temporary)
  sudo sysctl fs.inotify.max_user_watches=524288
  sudo sysctl fs.inotify.max_user_instances=512

  # Make permanent
  echo "fs.inotify.max_user_watches = 524288" | sudo tee -a /etc/sysctl.d/99-kind.conf
  echo "fs.inotify.max_user_instances = 512" | sudo tee -a /etc/sysctl.d/99-kind.conf
  sudo sysctl -p /etc/sysctl.d/99-kind.conf
  ```

- **Btrfs:**
  If you are using Btrfs, you might encounter issues with disk space or snapshotting. It is generally recommended to use `overlay2` storage driver for Docker.

## Troubleshooting

- **"command not found"**: Ensure `/usr/local/bin` is in your `$PATH`.
- **Docker permission denied**: Ensure your user is in the `docker` group (`sudo usermod -aG docker $USER`) or run with `sudo`.

## Next Steps

For a guide on common local use tasks such as deploying applications, port forwarding, and debugging, see [Local Use Tasks](local-use-tasks.md).
