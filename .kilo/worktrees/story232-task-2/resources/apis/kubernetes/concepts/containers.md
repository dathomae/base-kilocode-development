# Containers in Kubernetes

Containers are a technology for packaging an application along with its runtime dependencies.

## Why Containers?

Containers provide a consistent environment for applications to run in, regardless of the underlying infrastructure. This solves the "it works on my machine" problem.

## Container Images

A container image is a ready-to-run software package, containing everything needed to run an application: the code and any runtime it requires, application and system libraries, and default values for any essential settings.

By design, a container is immutable: you cannot change the code of a container that is already running. If you have a containerized application and want to make changes, you need to build a new image that includes your changes, then recreate the container to start from the updated image.

## Container Runtimes

Kubernetes supports several container runtimes:

- containerd
- CRI-O
- Docker Engine (using cri-dockerd)
- Mirantis Container Runtime

## Container Environment

Kubernetes provides a container environment that includes:

- **Filesystem**: A filesystem that is a combination of the container image and any volumes mounted into the container.
- **Container information**: Information about the container itself, such as its hostname and resources.
- **Cluster information**: Information about the cluster, such as the service account token and the API server address.
