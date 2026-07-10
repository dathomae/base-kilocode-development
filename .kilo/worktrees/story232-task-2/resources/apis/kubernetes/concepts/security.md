# Security

Kubernetes provides several mechanisms to secure your cluster and applications.

## Cloud Native Security

Security is a broad topic, and in the context of Kubernetes, it includes the security of the cluster itself, the security of the applications running on the cluster, and the security of the data stored in the cluster.

## Pod Security Standards

Pod Security Standards define three different policies to broadly cover the security spectrum. These standards let you define how you want to restrict the behavior of pods in a clear, consistent fashion.

## Service Accounts

A service account provides an identity for processes that run in a Pod.

## RBAC (Role-Based Access Control)

RBAC is a method of regulating access to computer or network resources based on the roles of individual users within your organization. RBAC authorization uses the `rbac.authorization.k8s.io` API group to drive authorization decisions, allowing you to dynamically configure policies through the Kubernetes API.
