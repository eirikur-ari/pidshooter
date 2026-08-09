FROM golang:1.26.5-trixie

ARG TZ
ENV TZ="$TZ"

ARG CLAUDE_CODE_VERSION=latest

RUN apt-get update && apt-get install -y --no-install-recommends \
    nodejs \
    npm \
    less \
    git \
    procps \
    sudo \
    fzf \
    zsh \
    man-db \
    unzip \
    gnupg2 \
    gh \
    iptables \
    ipset \
    iproute2 \
    dnsutils \
    aggregate \
    jq \
    vim \
    unzip \
    curl \
    wget \
    ca-certificates \
    fonts-firacode \
    locales \
    && rm -rf /var/lib/apt/lists/* \
    && sed -i 's/^# *\(en_US.UTF-8\)/\1/' /etc/locale.gen \
    && locale-gen en_US.UTF-8 \
    && update-locale LANG=en_US.UTF-8  \
    && apt-get autoremove -y \
    && apt-get clean -y

# Set locale environment variables
ENV LANG=en_US.UTF-8 \
    LC_ALL=en_US.UTF-8

# Set `DEVCONTAINER` environment variable to help with orientation
ENV DEVCONTAINER=true

ARG USERNAME=sandbox
ARG USER_UID=1000
ARG USER_GID=$USER_UID
RUN groupadd --gid $USER_GID $USERNAME \
    && useradd --uid $USER_UID --gid $USER_GID -m $USERNAME \
    && chsh -s /bin/zsh $USERNAME

RUN mkdir -p /workspace \
    /home/${USERNAME}/go/pkg/mod \
    /home/${USERNAME}/.cache/go-build \
    /home/${USERNAME}/.claude \
    /home/${USERNAME}/.config \
    && chown -R ${USERNAME}:${USERNAME}  /workspace /go /home/${USERNAME}

# Ensure default sandbox user has access to /usr/local/share
RUN mkdir -p /usr/local/share/npm-global \
    && chown -R ${USERNAME}:${USERNAME} /usr/local/share

WORKDIR /workspace

# Install Starship prompt
RUN curl -sS https://starship.rs/install.sh | sh -s -- -y
COPY --chown=${USERNAME}:${USERNAME} starship.toml /home/${USERNAME}/.config/starship.toml
RUN printf '#!/usr/bin/env zsh\neval "$(starship init zsh)"\n' > /home/${USERNAME}/.zshrc \
    && chown ${USERNAME}:${USERNAME} /home/${USERNAME}/.zshrc

# Copy and set up firewall script
COPY init-firewall.sh /usr/local/bin/
RUN chmod +x /usr/local/bin/init-firewall.sh && \
  echo "${USERNAME} ALL=(root) NOPASSWD: /usr/local/bin/init-firewall.sh" > /etc/sudoers.d/${USERNAME}-firewall && \
  chmod 0440 /etc/sudoers.d/${USERNAME}-firewall

# Set the default user for the container
USER $USERNAME

# Install global packages
ENV NPM_CONFIG_PREFIX=/usr/local/share/npm-global
ENV PATH=$PATH:/usr/local/share/npm-global/bin

# Install Claude Code CLI and set up configuration
RUN npm install -g @anthropic-ai/claude-code@${CLAUDE_CODE_VERSION} \
    && echo '{"hasCompletedOnboarding": true}' > /home/${USERNAME}/.claude/.claude.json
ENV CLAUDE_CONFIG_DIR=/home/${USERNAME}/.claude
ENV CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC=1
ENV DISABLE_AUTOUPDATER=1
ENV DO_NOT_TRACK=1

# Set environment variables for Go
ENV GOPATH=/home/${USERNAME}/go
ENV GOROOT=/usr/local/go
ENV PATH=$PATH:$GOPATH/bin:$GOROOT/bin
ENV GOMODCACHE=/home/${USERNAME}/go/pkg/mod

# Set the default shell to zsh rather than sh
ENV SHELL=/bin/zsh

# Set the default editor and visual
ENV EDITOR=vim
ENV VISUAL=vim
