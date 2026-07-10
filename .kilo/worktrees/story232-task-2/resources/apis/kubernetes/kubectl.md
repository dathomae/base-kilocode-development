# Installing and Using kubectl on Ubuntu Linux

`kubectl` is the Kubernetes command-line tool that allows you to run commands against Kubernetes clusters. This guide details how to install and configure `kubectl` specifically for Ubuntu Linux environments.

## Prerequisites

Before installing `kubectl`, ensure your system meets the following requirements:

- **Operating System:** Ubuntu (or other Debian-based distribution)
- **Architecture:** amd64, arm64, etc.
- **Privileges:** `sudo` access to install packages.
- **Dependencies:** The following packages are required for the installation process:
  - `apt-transport-https`
  - `ca-certificates`
  - `curl`
  - `gnupg`

## Installation Steps (Native Package Manager)

The recommended method for installing `kubectl` on Ubuntu is using the native `apt` package manager. This ensures easy updates and dependency management.

### 1. Update Package Index and Install Dependencies

First, update your `apt` package index and install the necessary packages to use the Kubernetes `apt` repository:

```bash
sudo apt-get update
sudo apt-get install -y apt-transport-https ca-certificates curl gnupg
```

### 2. Download the Public Signing Key

Download the public signing key for the Kubernetes package repositories.

```bash
# Create the keyrings directory if it doesn't exist
sudo mkdir -p -m 755 /etc/apt/keyrings

# Download the key and save it to the keyring location
curl -fsSL https://pkgs.k8s.io/core:/stable:/v1.35/deb/Release.key | sudo gpg --dearmor -o /etc/apt/keyrings/kubernetes-apt-keyring.gpg

# Set permissions for the key file
sudo chmod 644 /etc/apt/keyrings/kubernetes-apt-keyring.gpg
```

> **Note:** If you need a specific version other than `v1.35`, replace `v1.35` in the URL with your desired minor version (e.g., `v1.34`).

### 3. Add the Kubernetes apt Repository

Add the appropriate Kubernetes `apt` repository to your system's sources list.

```bash
# This overwrites any existing configuration in /etc/apt/sources.list.d/kubernetes.list
echo 'deb [signed-by=/etc/apt/keyrings/kubernetes-apt-keyring.gpg] https://pkgs.k8s.io/core:/stable:/v1.35/deb/ /' | sudo tee /etc/apt/sources.list.d/kubernetes.list

# Set permissions for the list file
sudo chmod 644 /etc/apt/sources.list.d/kubernetes.list
```

> **Note:** Ensure the version in the repository URL matches the version you used for the signing key.

### 4. Install kubectl

Update the `apt` package index again and install `kubectl`:

```bash
sudo apt-get update
sudo apt-get install -y kubectl
```

## Verification

After installation, verify that `kubectl` is installed correctly.

### Check Client Version

Run the following command to check the client version:

```bash
kubectl version --client
```

This should output the version details of the installed `kubectl` client.

### Check Cluster Connection (Optional)

If you have a Kubernetes cluster running and configured, you can check the connection:

```bash
kubectl cluster-info
```

If successful, this will display the URL of the Kubernetes control plane.

## Configuration Notes for Ubuntu

### Kubeconfig File

By default, `kubectl` looks for a configuration file named `config` in the `$HOME/.kube` directory.

- **Path:** `~/.kube/config`
- **Environment Variable:** You can specify a different kubeconfig file by setting the `KUBECONFIG` environment variable or by using the `--kubeconfig` flag.

### Shell Autocompletion

Enabling shell autocompletion is highly recommended for productivity.

**For Bash:**

1.  Install `bash-completion`:
    ```bash
    sudo apt-get install bash-completion
    ```
2.  Enable `kubectl` autocompletion:
    ```bash
    echo 'source <(kubectl completion bash)' >> ~/.bashrc
    ```
3.  (Optional) If you use an alias like `k=kubectl`:
    ```bash
    echo 'alias k=kubectl' >> ~/.bashrc
    echo 'complete -o default -F __start_kubectl k' >> ~/.bashrc
    ```
4.  Reload your shell:
    ```bash
    source ~/.bashrc
    ```

**For Zsh:**

If you are using Zsh on Ubuntu:

```bash
echo 'source <(kubectl completion zsh)' >> ~/.zshrc
source ~/.zshrc
```

## Troubleshooting

- **"No Auth Provider Found":** If you encounter this error with cloud providers (AWS, GKE, Azure), you may need to install specific authentication plugins (e.g., `aws-iam-authenticator`, `gke-gcloud-auth-plugin`).
- **Connection Refused:** Ensure your cluster is running and your `~/.kube/config` file is correctly pointing to the cluster's API server.

## Common Tasks

For a guide on common local use tasks using `kubectl` with `kind`, such as deploying applications, port forwarding, and debugging, see [Local Use Tasks](local-use-tasks.md).
