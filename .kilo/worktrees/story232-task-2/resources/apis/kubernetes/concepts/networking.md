# Services, Load Balancing, and Networking

Kubernetes networking addresses four concerns:

1.  Highly-coupled container-to-container communications: this is solved by Pods and `localhost` communications.
2.  Pod-to-Pod communications: this is the primary focus of this document.
3.  Pod-to-Service communications: this is covered by Services.
4.  External-to-Service communications: this is covered by Services.

## Service

An abstract way to expose an application running on a set of Pods as a network service. With Kubernetes you don't need to modify your application to use an unfamiliar service discovery mechanism. Kubernetes gives Pods their own IP addresses and a single DNS name for a set of Pods, and can load-balance across them.

## Ingress

An API object that manages external access to the services in a cluster, typically HTTP. Ingress may provide load balancing, SSL termination and name-based virtual hosting.

## Network Policies

NetworkPolicies are an application-centric construct that allow you to specify how a pod is allowed to communicate with various network "entities" (we use the word "entity" here to avoid overloading the more common terms such as "endpoints" and "services", which have specific Kubernetes meanings) over the network.
