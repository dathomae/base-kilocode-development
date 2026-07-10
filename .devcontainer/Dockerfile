FROM ubuntu:24.04

ENV DEBIAN_FRONTEND=noninteractive

# Common tools + Python
RUN apt-get update && apt-get install -y --no-install-recommends \
    curl wget git build-essential util-linux \
    python3.12 python3.12-venv python3.12-dev \
    ca-certificates vim emacs-nox jq libpq-dev \
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
