# Local Use Tasks with Kind and Kubectl on Ubuntu

This guide provides practical examples and explanations for common tasks performed during local use (such as demos or development) with a single Kubernetes cluster on a laptop, using `kind` and `kubectl`.

## Cluster Management

### Quick Cluster Creation

For most local use cases, a default cluster is sufficient.

```bash
kind create cluster --name demo-cluster
```

### Checking Cluster Status

Verify that your cluster is running and `kubectl` is configured to talk to it.

```bash
kubectl cluster-info --context kind-demo-cluster
```

### Switching Contexts

If you work with multiple clusters (e.g., `kind`, `minikube`, or remote clusters), you need to switch contexts.

```bash
# List available contexts
kubectl config get-contexts

# Switch to the kind cluster
kubectl config use-context kind-demo-cluster
```

### Deleting the Cluster

When you are done, delete the cluster to free up resources (Docker containers).

```bash
kind delete cluster --name demo-cluster
```

## Application Deployment

### Deploying a Simple Application

Deploy a simple web server (e.g., NGINX) to test your cluster.

**Imperative Method (Quick testing):**

```bash
kubectl create deployment nginx-app --image=nginx:latest
```

**Declarative Method (Recommended for reproducibility):**
Create a file named `nginx-deployment.yaml`:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-app
spec:
  replicas: 1
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
        - name: nginx
          image: nginx:latest
          ports:
            - containerPort: 80
```

Apply the configuration:

```bash
kubectl apply -f nginx-deployment.yaml
```

### Verifying Deployment

Check if the Pods are running.

```bash
kubectl get pods
# Output should show status 'Running'
```

## Accessing Applications

In a local `kind` environment, Pods are running inside Docker containers and are not directly accessible via `localhost` by default.

### Port Forwarding (Easiest Method)

The simplest way to access a service or pod from your laptop is using `kubectl port-forward`. This forwards a port on your local machine to a port on the pod.

1.  **Forward a Pod port:**

    ```bash
    # Syntax: kubectl port-forward <pod-name> <local-port>:<pod-port>
    kubectl port-forward deployment/nginx-app 8080:80
    ```

    Now, access the app at `http://localhost:8080`.

2.  **Forward a Service port:**
    First, expose the deployment as a service:
    ```bash
    kubectl expose deployment nginx-app --port=80 --target-port=80 --type=ClusterIP
    ```
    Then forward to the service:
    ```bash
    kubectl port-forward service/nginx-app 8080:80
    ```

### NodePort (Alternative)

You can expose a service on a specific port of the `kind` node (Docker container).

```bash
kubectl expose deployment nginx-app --type=NodePort --name=nginx-nodeport --port=80
```

To access it, you need the node's IP and the assigned NodePort.

```bash
# Get the NodePort (e.g., 31234)
kubectl get svc nginx-nodeport

# Get the Node IP (usually the kind container IP)
kubectl get nodes -o wide
```

_Note: Port forwarding is generally preferred for simple local testing._

## Debugging and Troubleshooting

### Viewing Logs

Check the standard output/error of a container.

```bash
# Follow logs (-f) for a specific pod
kubectl logs -f <pod-name>

# If a pod has multiple containers, specify the container name
kubectl logs <pod-name> -c <container-name>

# View logs for all pods in a deployment (using label selector)
kubectl logs -f -l app=nginx
```

### Describing Resources

Get detailed information about a resource, including events (errors, scheduling issues).

```bash
kubectl describe pod <pod-name>
kubectl describe deployment nginx-app
```

_Look at the "Events" section at the bottom for clues about failures (e.g., ImagePullBackOff, CrashLoopBackOff)._

### Executing Commands in a Container

Get a shell inside a running container to inspect files or run diagnostic commands.

```bash
kubectl exec -it <pod-name> -- /bin/bash
# Or /bin/sh if bash is not available
```

### Debugging CrashLoopBackOff

If a pod keeps crashing:

1.  Check `kubectl get pods` to see the restart count.
2.  Check `kubectl logs <pod-name> --previous` to see logs from the _crashed_ instance.
3.  Check `kubectl describe pod <pod-name>` for exit codes and reasons.

## Resource Management

### Scaling Applications

Manually scale the number of replicas.

```bash
kubectl scale deployment nginx-app --replicas=3
```

Verify with `kubectl get pods`.

### Editing Resources Live

Edit the configuration of a running resource. This opens the manifest in your default editor (e.g., `vi` or `nano`).

```bash
kubectl edit deployment nginx-app
```

_Changes are applied immediately upon saving and exiting._

### Deleting Resources

Remove deployments, services, or pods.

```bash
# Delete by filename (declarative)
kubectl delete -f nginx-deployment.yaml

# Delete by name (imperative)
kubectl delete service nginx-app
kubectl delete deployment nginx-app
```

## Configuration Management

### Using ConfigMaps

Inject configuration data into pods.

1.  **Create a ConfigMap:**

    ```bash
    kubectl create configmap app-config --from-literal=key1=value1 --from-literal=key2=value2
    ```

2.  **Use in a Pod:**
    Modify your deployment YAML to use the ConfigMap as environment variables.
    ```yaml
    envFrom:
      - configMapRef:
          name: app-config
    ```

### Using Secrets

Store sensitive data (passwords, keys).

1.  **Create a Secret:**

    ```bash
    kubectl create secret generic app-secret --from-literal=password=my-secret-password
    ```

2.  **Use in a Pod:**
    Similar to ConfigMaps, mount as environment variables or volumes.

## Advanced Local Setup

### Loading Local Images into Kind

If you build a Docker image locally, `kind` won't see it by default. You must load it into the cluster nodes.

1.  **Build the image:**

    ```bash
    docker build -t my-app:local .
    ```

2.  **Load into Kind:**

    ```bash
    kind load docker-image my-app:local --name demo-cluster
    ```

3.  **Use in Deployment:**
    Ensure your deployment uses `imagePullPolicy: IfNotPresent` or `Never` so it uses the loaded image instead of trying to pull from a registry.
    ```yaml
    containers:
      - name: my-app
        image: my-app:local
        imagePullPolicy: IfNotPresent
    ```
