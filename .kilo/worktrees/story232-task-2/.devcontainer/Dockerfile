FROM ubuntu:24.04

# Prevent interactive prompts during apt operations
ENV DEBIAN_FRONTEND=noninteractive

# Update apt and install system dependencies, Python 3.12, and build tools
RUN apt-get update && apt-get install -y --no-install-recommends \
    curl \
    wget \
    git \
    build-essential \
    util-linux \
    python3.12 \
    python3.12-venv \
    python3.12-dev \
    libpq-dev \
    ca-certificates \
    # Editors
    vim \
    emacs-nox \
    # Network debugging tools
    iputils-ping \
    dnsutils \
    net-tools \
    traceroute \
    netcat-openbsd \
    iproute2 \
    tcpdump \
    nmap \
    openssh-client \
    # Frequently used by AI to extract stuff from JSON
    jq \
    # java, for antlr
    openjdk-21-jdk \
    xxd \
    # Used for example expected results capture and verification
    libreoffice \
    && rm -rf /var/lib/apt/lists/*

# Install Node.js 20.x, npm, and pnpm
RUN curl -fsSL https://deb.nodesource.com/setup_20.x | bash - \
    && apt-get install -y nodejs \
    && npm install -g pnpm \
    && rm -rf /var/lib/apt/lists/*

# Install GitHub CLI
RUN curl -fsSL https://cli.github.com/packages/githubcli-archive-keyring.gpg \
    -o /usr/share/keyrings/githubcli-archive-keyring.gpg \
    && echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" \
    > /etc/apt/sources.list.d/github-cli.list \
    && apt-get update \
    && apt-get install -y --no-install-recommends gh \
    && rm -rf /var/lib/apt/lists/*

# Install uv (Python package manager written in Rust)
RUN curl -LsSf https://astral.sh/uv/install.sh | env UV_INSTALL_DIR="/usr/local/bin" sh

# Set up the workspace
RUN mkdir -p /${LOCAL_WORKSPACE_FOLDER_NAME}
WORKDIR /${LOCAL_WORKSPACE_FOLDER_NAME}


