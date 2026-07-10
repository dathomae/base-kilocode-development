# Initialize Project

This script bootstraps a new project with language-specific source scaffolding, a Docker devcontainer configuration, a full memory bank structure, and placeholder requirements. It is the single entry point for starting a new KiloCode-managed project.

## Parameters

| Name | Type | Description | Required | Default Value |
|------|------|-------------|----------|---------------|
| project_name | string | The name of the project (used for package/module names, brief.md, etc.) | Yes | N/A |
| language | string | One of: `go`, `python`, `java`, `kotlin`, `rust`, `c` | Yes | N/A |
| description | string | Brief description of the project's purpose (1-3 sentences) | Yes | N/A |
| requirements | string | Requirements text or path to a requirements file | No | (empty — creates TBD placeholder) |
| project_path | string | Path where the project will be created | No | Current working directory |

## Operations

### Phase 1: Validation and Defaults

1. Validate that all required parameters are provided and non-empty.
2. Validate that `language` is one of the supported values: `go`, `python`, `java`, `kotlin`, `rust`, `c`.
3. If `project_path` is not specified, use the current working directory.
4. Determine the package/module path based on language:
   - **Go**: Use `github.com/<project_name>` as the module path. Confirm with the user.
   - **Python**: Use `<project_name>` (snake_case conversion of project_name) as the package name.
   - **Java**: Use `com.<project_name>` as the base package. Confirm with the user.
   - **Kotlin**: Use `com.<project_name>` as the base package. Confirm with the user.
   - **Rust**: Cargo handles naming; use project_name as-is.
   - **C**: No package concept; use project_name as the project directory identifier.
5. Announce the chosen defaults to the user and ask for confirmation before proceeding:
   ```
   Project initialization defaults for <project_name> (<language>):
     Source root:       src/<package_name>/
     Build system:      <build_system>
     Test framework:    <test_framework>
     Dev environment:   Docker devcontainer (Ubuntu 24.04)
   
   Proceed with these defaults? (y/n)
   ```

### Phase 2: Source Directory Scaffolding

6. Create the language-specific source directory structure:

   **Go:**
   a. Create `src/` directory.
   b. Create `src/<project_name>/` package directory.
   c. Create `src/<project_name>/go.mod` with module path `github.com/<project_name>` and Go version matching `go version` output.
   d. Create `src/<project_name>/main.go` stub:
      ```go
      package main
      
      import "fmt"
      
      func main() {
          fmt.Println("<project_name>")
      }
      ```
   e. Create `src/<project_name>/main_test.go` stub:
      ```go
      package main
      
      import "testing"
      
      func TestMain(t *testing.T) {
          // TODO: Add tests
      }
      ```

   **Python:**
   a. Create `src/<project_name>/` package directory.
   b. Create `src/<project_name>/__init__.py` (empty file).
   c. Create `pyproject.toml` with setuptools configuration:
      ```toml
      [build-system]
      requires = ["setuptools>=61.0"]
      build-backend = "setuptools.backends._legacy:_Backend"
      
      [project]
      name = "<project_name>"
      version = "0.1.0"
      description = "<description>"
      requires-python = ">=3.12"
      dependencies = []
      
      [project.optional-dependencies]
      dev = ["pytest", "pytest-cov"]
      
      [tool.pytest.ini_options]
      testpaths = ["tests"]
      ```
   d. Create `tests/` directory.
   e. Create `tests/__init__.py` (empty file).
   f. Create `tests/test_main.py` stub:
      ```python
      """Tests for <project_name>."""
      
      def test_placeholder():
          """TODO: Add real tests."""
          pass
      ```

   **Java:**
   a. Create `src/main/java/com/<project_name>/` package directory.
   b. Create `src/test/java/com/<project_name>/` package directory.
   c. Create `src/main/java/com/<project_name>/Main.java` stub:
      ```java
      package com.<project_name>;
      
      public class Main {
          public static void main(String[] args) {
              System.out.println("<project_name>");
          }
      }
      ```
   d. Create `src/test/java/com/<project_name>/MainTest.java` stub:
      ```java
      package com.<project_name>;
      
      import org.junit.jupiter.api.Test;
      import static org.junit.jupiter.api.Assertions.*;
      
      class MainTest {
          @Test
          void testPlaceholder() {
              // TODO: Add real tests
              assertTrue(true);
          }
      }
      ```
   e. Create `build.gradle.kts`:
      ```kotlin
      plugins {
          java
      }
      
      group = "com.<project_name>"
      version = "0.1.0"
      
      repositories {
          mavenCentral()
      }
      
      dependencies {
          testImplementation("org.junit.jupiter:junit-jupiter:5.11.0")
      }
      
      java {
          toolchain {
              languageVersion = JavaLanguageVersion.of(21)
          }
      }
      
      tasks.test {
          useJUnitPlatform()
      }
      ```
   f. Create `settings.gradle.kts`:
      ```kotlin
      rootProject.name = "<project_name>"
      ```

   **Kotlin:**
   a. Create `src/main/kotlin/com/<project_name>/` package directory.
   b. Create `src/test/kotlin/com/<project_name>/` package directory.
   c. Create `src/main/kotlin/com/<project_name>/Main.kt` stub:
      ```kotlin
      package com.<project_name>
      
      fun main() {
          println("<project_name>")
      }
      ```
   d. Create `src/test/kotlin/com/<project_name>/MainTest.kt` stub:
      ```kotlin
      package com.<project_name>
      
      import kotlin.test.Test
      import kotlin.test.assertTrue
      
      class MainTest {
          @Test
          fun testPlaceholder() {
              // TODO: Add real tests
              assertTrue(true)
          }
      }
      ```
   e. Create `build.gradle.kts`:
      ```kotlin
      plugins {
          kotlin("jvm") version "2.1.0"
      }
      
      group = "com.<project_name>"
      version = "0.1.0"
      
      repositories {
          mavenCentral()
      }
      
      dependencies {
          testImplementation(kotlin("test"))
      }
      
      kotlin {
          jvmToolchain(21)
      }
      
      tasks.test {
          useJUnitPlatform()
      }
      ```
   f. Create `settings.gradle.kts`:
      ```kotlin
      rootProject.name = "<project_name>"
      ```

   **Rust:**
   a. Create `src/` directory.
   b. Create `src/main.rs` stub:
      ```rust
      fn main() {
          println!("<project_name>");
      }
      ```
   c. Create `Cargo.toml`:
      ```toml
      [package]
      name = "<project_name>"
      version = "0.1.0"
      edition = "2021"
      description = "<description>"
      
      [dependencies]
      
      [dev-dependencies]
      ```
   d. Create `tests/` directory.
   e. Create `tests/integration_test.rs` stub:
      ```rust
      #[test]
      fn test_placeholder() {
          // TODO: Add real tests
          assert!(true);
      }
      ```

   **C:**
   a. Create `src/` directory.
   b. Create `src/main.c` stub:
      ```c
      #include <stdio.h>
      
      int main(int argc, char *argv[]) {
          printf("<project_name>\n");
          return 0;
      }
      ```
   c. Create `include/` directory.
   d. Create `tests/` directory.
   e. Create `tests/test_main.c` stub:
      ```c
      #include <assert.h>
      
      void test_placeholder() {
          // TODO: Add real tests
          assert(1);
      }
      
      int main() {
          test_placeholder();
          return 0;
      }
      ```
   f. Create `Makefile`:
      ```makefile
      CC = gcc
      CFLAGS = -Wall -Wextra -std=c11 -Iinclude
      TARGET = <project_name>
      SRCDIR = src
      OBJDIR = build
      TESTDIR = tests
      
      SRCS = $(wildcard $(SRCDIR)/*.c)
      OBJS = $(SRCS:$(SRCDIR)/%.c=$(OBJDIR)/%.o)
      
      .PHONY: all clean test
      
      all: $(TARGET)
      
      $(TARGET): $(OBJS)
      	$(CC) $(CFLAGS) -o $@ $^
      
      $(OBJDIR)/%.o: $(SRCDIR)/%.c | $(OBJDIR)
      	$(CC) $(CFLAGS) -c -o $@ $<
      
      $(OBJDIR):
      	mkdir -p $(OBJDIR)
      
      test: $(TARGET)
      	$(CC) $(CFLAGS) -o $(OBJDIR)/test_main $(TESTDIR)/test_main.c
      	./$(OBJDIR)/test_main
      
      clean:
      	rm -rf $(OBJDIR) $(TARGET)
      ```

### Phase 3: Devcontainer Setup

7. Create the `.devcontainer/` directory.

8. Create `.devcontainer/.dockerignore`:
   ```
   .git
   .gitignore
   __pycache__
   .dockerignore
   Dockerfile
   docker-compose.yml
   README.md
   ```

9. Create `.devcontainer/Dockerfile` with language-specific toolchain. All Dockerfiles start from `FROM ubuntu:24.04` and include common tools: `curl`, `wget`, `git`, `build-essential`, `vim`, `emacs-nox`, `jq`, `iputils-ping`, `dnsutils`, `net-tools`, `traceroute`, `netcat-openbsd`, `iproute2`, `tcpdump`, `nmap`, `openssh-client`, `xxd`, plus Node.js 20 with pnpm, and GitHub CLI. Then add language-specific toolchain:

   **Go:**
   ```dockerfile
   FROM ubuntu:24.04
   
   ENV DEBIAN_FRONTEND=noninteractive
   
   # Common tools
   RUN apt-get update && apt-get install -y --no-install-recommends \
       curl wget git build-essential util-linux \
       ca-certificates vim emacs-nox jq \
       iputils-ping dnsutils net-tools traceroute netcat-openbsd iproute2 tcpdump nmap openssh-client xxd \
       && rm -rf /var/lib/apt/lists/*
   
   # Node.js and pnpm
   RUN curl -fsSL https://deb.nodesource.com/setup_20.x | bash - \
       && apt-get install -y nodejs \
       && npm install -g pnpm \
       && rm -rf /var/lib/apt/lists/*
   
   # GitHub CLI
   RUN curl -fsSL https://cli.github.com/packages/githubcli-archive-keyring.gpg \
       -o /usr/share/keyrings/githubcli-archive-keyring.gpg \
       && echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" \
       > /etc/apt/sources.list.d/github-cli.list \
       && apt-get update \
       && apt-get install -y --no-install-recommends gh \
       && rm -rf /var/lib/apt/lists/*
   
   # Go
   ENV GO_VERSION=1.24.0
   RUN curl -sSL "https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz" | tar -xz -C /usr/local
   ENV PATH=$PATH:/usr/local/go/bin
   
   # Workspace
   RUN mkdir -p /${LOCAL_WORKSPACE_FOLDER_NAME}
   WORKDIR /${LOCAL_WORKSPACE_FOLDER_NAME}
   ```

   **Python:**
   ```dockerfile
   FROM ubuntu:24.04
   
   ENV DEBIAN_FRONTEND=noninteractive
   
   # Common tools + Python
   RUN apt-get update && apt-get install -y --no-install-recommends \
       curl wget git build-essential util-linux \
       python3.12 python3.12-venv python3.12-dev \
       ca-certificates vim emacs-nox jq \
       iputils-ping dnsutils net-tools traceroute netcat-openbsd iproute2 tcpdump nmap openssh-client xxd \
       && rm -rf /var/lib/apt/lists/*
   
   # Node.js and pnpm
   RUN curl -fsSL https://deb.nodesource.com/setup_20.x | bash - \
       && apt-get install -y nodejs \
       && npm install -g pnpm \
       && rm -rf /var/lib/apt/lists/*
   
   # GitHub CLI
   RUN curl -fsSL https://cli.github.com/packages/githubcli-archive-keyring.gpg \
       -o /usr/share/keyrings/githubcli-archive-keyring.gpg \
       && echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" \
       > /etc/apt/sources.list.d/github-cli.list \
       && apt-get update \
       && apt-get install -y --no-install-recommends gh \
       && rm -rf /var/lib/apt/lists/*
   
   # uv (Python package manager)
   RUN curl -LsSf https://astral.sh/uv/install.sh | env UV_INSTALL_DIR="/usr/local/bin" sh
   
   # Workspace
   RUN mkdir -p /${LOCAL_WORKSPACE_FOLDER_NAME}
   WORKDIR /${LOCAL_WORKSPACE_FOLDER_NAME}
   ```

   **Java / Kotlin:**
   ```dockerfile
   FROM ubuntu:24.04
   
   ENV DEBIAN_FRONTEND=noninteractive
   
   # Common tools + Java
   RUN apt-get update && apt-get install -y --no-install-recommends \
       curl wget git build-essential util-linux \
       openjdk-21-jdk \
       ca-certificates vim emacs-nox jq \
       iputils-ping dnsutils net-tools traceroute netcat-openbsd iproute2 tcpdump nmap openssh-client xxd \
       && rm -rf /var/lib/apt/lists/*
   
   # Node.js and pnpm
   RUN curl -fsSL https://deb.nodesource.com/setup_20.x | bash - \
       && apt-get install -y nodejs \
       && npm install -g pnpm \
       && rm -rf /var/lib/apt/lists/*
   
   # GitHub CLI
   RUN curl -fsSL https://cli.github.com/packages/githubcli-archive-keyring.gpg \
       -o /usr/share/keyrings/githubcli-archive-keyring.gpg \
       && echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" \
       > /etc/apt/sources.list.d/github-cli.list \
       && apt-get update \
       && apt-get install -y --no-install-recommends gh \
       && rm -rf /var/lib/apt/lists/*
   
   # Workspace
   RUN mkdir -p /${LOCAL_WORKSPACE_FOLDER_NAME}
   WORKDIR /${LOCAL_WORKSPACE_FOLDER_NAME}
   ```

   **Rust:**
   ```dockerfile
   FROM ubuntu:24.04
   
   ENV DEBIAN_FRONTEND=noninteractive
   
   # Common tools
   RUN apt-get update && apt-get install -y --no-install-recommends \
       curl wget git build-essential util-linux \
       ca-certificates vim emacs-nox jq \
       iputils-ping dnsutils net-tools traceroute netcat-openbsd iproute2 tcpdump nmap openssh-client xxd \
       && rm -rf /var/lib/apt/lists/*
   
   # Node.js and pnpm
   RUN curl -fsSL https://deb.nodesource.com/setup_20.x | bash - \
       && apt-get install -y nodejs \
       && npm install -g pnpm \
       && rm -rf /var/lib/apt/lists/*
   
   # GitHub CLI
   RUN curl -fsSL https://cli.github.com/packages/githubcli-archive-keyring.gpg \
       -o /usr/share/keyrings/githubcli-archive-keyring.gpg \
       && echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" \
       > /etc/apt/sources.list.d/github-cli.list \
       && apt-get update \
       && apt-get install -y --no-install-recommends gh \
       && rm -rf /var/lib/apt/lists/*
   
   # Rust
   RUN curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y
   ENV PATH=$PATH:/root/.cargo/bin
   
   # Workspace
   RUN mkdir -p /${LOCAL_WORKSPACE_FOLDER_NAME}
   WORKDIR /${LOCAL_WORKSPACE_FOLDER_NAME}
   ```

   **C:**
   ```dockerfile
   FROM ubuntu:24.04
   
   ENV DEBIAN_FRONTEND=noninteractive
   
   # Common tools + C tooling
   RUN apt-get update && apt-get install -y --no-install-recommends \
       curl wget git build-essential gdb make cmake util-linux \
       ca-certificates vim emacs-nox jq \
       iputils-ping dnsutils net-tools traceroute netcat-openbsd iproute2 tcpdump nmap openssh-client xxd \
       && rm -rf /var/lib/apt/lists/*
   
   # Node.js and pnpm
   RUN curl -fsSL https://deb.nodesource.com/setup_20.x | bash - \
       && apt-get install -y nodejs \
       && npm install -g pnpm \
       && rm -rf /var/lib/apt/lists/*
   
   # GitHub CLI
   RUN curl -fsSL https://cli.github.com/packages/githubcli-archive-keyring.gpg \
       -o /usr/share/keyrings/githubcli-archive-keyring.gpg \
       && echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" \
       > /etc/apt/sources.list.d/github-cli.list \
       && apt-get update \
       && apt-get install -y --no-install-recommends gh \
       && rm -rf /var/lib/apt/lists/*
   
   # Workspace
   RUN mkdir -p /${LOCAL_WORKSPACE_FOLDER_NAME}
   WORKDIR /${LOCAL_WORKSPACE_FOLDER_NAME}
   ```

10. Create `.devcontainer/docker-compose.yml`:
    ```yaml
    services:
      app:
        extra_hosts:
          - "host.docker.internal:host-gateway"
        build:
          context: ..
          dockerfile: .devcontainer/Dockerfile
        command: sleep infinity
        volumes:
          - ..:/${LOCAL_WORKSPACE_FOLDER_NAME}:cached
        networks:
          - dev-network
    
    networks:
      dev-network:
    ```
    Note: No database or other services by default. Users add services as their project requires.

11. Create `.devcontainer/devcontainer.json` with language-specific VS Code extensions:

    **Go:**
    ```json
    {
        "name": "<project_name> Dev Container",
        "dockerComposeFile": "docker-compose.yml",
        "service": "app",
        "workspaceFolder": "/${localWorkspaceFolderBasename}",
        "initializeCommand": "echo LOCAL_WORKSPACE_FOLDER_NAME=${localWorkspaceFolderBasename} > .devcontainer/.env",
        "mounts": [
            "source=${env:HOME}/.gitconfig,target=/root/.gitconfig,type=bind,consistency=cached"
        ],
        "customizations": {
            "vscode": {
                "extensions": [
                    "golang.go",
                    "gururagavendra.local-pr-review"
                ]
            }
        },
        "postCreateCommand": "bash .devcontainer/post-create.sh /${localWorkspaceFolderBasename}",
        "remoteUser": "root"
    }
    ```

    **Python:**
    ```json
    {
        "name": "<project_name> Dev Container",
        "dockerComposeFile": "docker-compose.yml",
        "service": "app",
        "workspaceFolder": "/${localWorkspaceFolderBasename}",
        "initializeCommand": "echo LOCAL_WORKSPACE_FOLDER_NAME=${localWorkspaceFolderBasename} > .devcontainer/.env",
        "mounts": [
            "source=${env:HOME}/.gitconfig,target=/root/.gitconfig,type=bind,consistency=cached"
        ],
        "customizations": {
            "vscode": {
                "extensions": [
                    "ms-python.python",
                    "ms-python.debugpy",
                    "gururagavendra.local-pr-review"
                ]
            }
        },
        "postCreateCommand": "bash .devcontainer/post-create.sh /${localWorkspaceFolderBasename}",
        "remoteUser": "root"
    }
    ```

    **Java:**
    ```json
    {
        "name": "<project_name> Dev Container",
        "dockerComposeFile": "docker-compose.yml",
        "service": "app",
        "workspaceFolder": "/${localWorkspaceFolderBasename}",
        "initializeCommand": "echo LOCAL_WORKSPACE_FOLDER_NAME=${localWorkspaceFolderBasename} > .devcontainer/.env",
        "mounts": [
            "source=${env:HOME}/.gitconfig,target=/root/.gitconfig,type=bind,consistency=cached"
        ],
        "customizations": {
            "vscode": {
                "extensions": [
                    "vscjava.vscode-java-pack",
                    "gururagavendra.local-pr-review"
                ]
            }
        },
        "postCreateCommand": "bash .devcontainer/post-create.sh /${localWorkspaceFolderBasename}",
        "remoteUser": "root"
    }
    ```

    **Kotlin:**
    ```json
    {
        "name": "<project_name> Dev Container",
        "dockerComposeFile": "docker-compose.yml",
        "service": "app",
        "workspaceFolder": "/${localWorkspaceFolderBasename}",
        "initializeCommand": "echo LOCAL_WORKSPACE_FOLDER_NAME=${localWorkspaceFolderBasename} > .devcontainer/.env",
        "mounts": [
            "source=${env:HOME}/.gitconfig,target=/root/.gitconfig,type=bind,consistency=cached"
        ],
        "customizations": {
            "vscode": {
                "extensions": [
                    "fwcd.kotlin",
                    "vscjava.vscode-java-pack",
                    "gururagavendra.local-pr-review"
                ]
            }
        },
        "postCreateCommand": "bash .devcontainer/post-create.sh /${localWorkspaceFolderBasename}",
        "remoteUser": "root"
    }
    ```

    **Rust:**
    ```json
    {
        "name": "<project_name> Dev Container",
        "dockerComposeFile": "docker-compose.yml",
        "service": "app",
        "workspaceFolder": "/${localWorkspaceFolderBasename}",
        "initializeCommand": "echo LOCAL_WORKSPACE_FOLDER_NAME=${localWorkspaceFolderBasename} > .devcontainer/.env",
        "mounts": [
            "source=${env:HOME}/.gitconfig,target=/root/.gitconfig,type=bind,consistency=cached"
        ],
        "customizations": {
            "vscode": {
                "extensions": [
                    "rust-lang.rust-analyzer",
                    "gururagavendra.local-pr-review"
                ]
            }
        },
        "postCreateCommand": "bash .devcontainer/post-create.sh /${localWorkspaceFolderBasename}",
        "remoteUser": "root"
    }
    ```

    **C:**
    ```json
    {
        "name": "<project_name> Dev Container",
        "dockerComposeFile": "docker-compose.yml",
        "service": "app",
        "workspaceFolder": "/${localWorkspaceFolderBasename}",
        "initializeCommand": "echo LOCAL_WORKSPACE_FOLDER_NAME=${localWorkspaceFolderBasename} > .devcontainer/.env",
        "mounts": [
            "source=${env:HOME}/.gitconfig,target=/root/.gitconfig,type=bind,consistency=cached"
        ],
        "customizations": {
            "vscode": {
                "extensions": [
                    "ms-vscode.cpptools",
                    "gururagavendra.local-pr-review"
                ]
            }
        },
        "postCreateCommand": "bash .devcontainer/post-create.sh /${localWorkspaceFolderBasename}",
        "remoteUser": "root"
    }
    ```

12. Create `.devcontainer/post-create.sh` with language-specific bootstrap:

    **Go:**
    ```bash
    #!/bin/bash
    set -e
    echo "=== Bootstrapping Go Development Environment ==="
    cd "$1"
    /usr/local/go/bin/go mod download
    echo "=== Bootstrap Complete! ==="
    ```

    **Python:**
    ```bash
    #!/bin/bash
    set -e
    echo "=== Bootstrapping Python Development Environment ==="
    cd "$1"
    uv sync
    echo "=== Bootstrap Complete! ==="
    ```

    **Java:**
    ```bash
    #!/bin/bash
    set -e
    echo "=== Bootstrapping Java Development Environment ==="
    cd "$1"
    ./gradlew build
    echo "=== Bootstrap Complete! ==="
    ```

    **Kotlin:**
    ```bash
    #!/bin/bash
    set -e
    echo "=== Bootstrapping Kotlin Development Environment ==="
    cd "$1"
    ./gradlew build
    echo "=== Bootstrap Complete! ==="
    ```

    **Rust:**
    ```bash
    #!/bin/bash
    set -e
    echo "=== Bootstrapping Rust Development Environment ==="
    cd "$1"
    cargo build
    echo "=== Bootstrap Complete! ==="
    ```

    **C:**
    ```bash
    #!/bin/bash
    set -e
    echo "=== Bootstrapping C Development Environment ==="
    cd "$1"
    make
    echo "=== Bootstrap Complete! ==="
    ```

13. Make `.devcontainer/post-create.sh` executable (`chmod +x`).

### Phase 4: Memory Bank Initialization

14. Execute the `initialize-per-project-memory-bank.pdd.script.md` script with:
    - `project_path` set to the target project path
    - `project_name` set to the provided project_name
    - `project_description` set to the provided description

15. Populate `memory-bank/brief.md` with the user-provided description.

16. If the user provided requirements (text or file path):
    - Create `memory-bank/specifications/requirements/<project_name>_requirements.md` with the provided content.
    - Add an entry to `memory-bank/specifications/requirements/toc.md` pointing to the file.
    If no requirements were provided:
    - Create `memory-bank/specifications/requirements/<project_name>_requirements.md` with:
      ```markdown
      # <project_name> Requirements
      
      ## Status: TBD
      
      Requirements have not yet been defined. This document will be populated as requirements are elaborated.
      ```
    - Add an entry to `memory-bank/specifications/requirements/toc.md` pointing to the file with a "TBD" status note.

### Phase 5: Final Report

17. Report a summary of everything created:
    - Source directory structure (list of files created)
    - Devcontainer files created
    - Memory bank directories and files created
    - Requirements document location and status
    - The project is ready for development

## Output

The script outputs:
- Confirmation of all files and directories created, organized by phase
- The language and ecosystem defaults that were used
- The path to the initialized project
- Instructions for next steps (how to open in devcontainer, how to run tests, how to create first story)

## Error Handling

If required parameters aren't specified, the user should be prompted. The operations in the script WILL NOT be executed until/unless all required parameters have been specified. During execution, if an error is encountered (e.g., permission denied, language toolchain not found), the execution of the script will be paused and the user will be asked if script execution should continue or if the script should be terminated.

Specific error conditions handled:
- Missing required parameters
- Invalid language value
- Project path already contains files (warn user, ask to confirm overwrite)
- Permission denied when creating directories
- Memory bank initialization script not found
