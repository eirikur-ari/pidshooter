FROM golang:1.26.5-trixie

ARG CLAUDE_CODE_VERSION=2.1.212

RUN apt-get update && apt-get install -y --no-install-recommends \
    nodejs \
    npm \
    git \
    ca-certificates \
    fzf \
    zsh \
    jq \
    vim \
    unzip \
    curl \
    wget \
    fonts-firacode \
    && apt-get clean -y && rm -rf /var/lib/apt/lists/* \
    && npm install -g @anthropic-ai/claude-code@${CLAUDE_CODE_VERSION}

ARG USERNAME=sandbox
ARG USER_UID=1000
ARG USER_GID=$USER_UID
RUN groupadd --gid $USER_GID $USERNAME \
    && useradd --uid $USER_UID --gid $USER_GID -m $USERNAME \
    && chsh -s /bin/zsh $USERNAME

RUN mkdir -p /home/${USERNAME}/go/pkg/mod \
    && mkdir -p /home/${USERNAME}/.cache/go-build \
    && mkdir -p /home/${USERNAME}/.claude \
    && mkdir -p /home/${USERNAME}/.config \
    && echo '{"hasCompletedOnboarding": true}' > /home/${USERNAME}/.claude/.claude.json \
    && chown -R ${USERNAME}:${USERNAME} /home/${USERNAME}

# Install Starship prompt
RUN curl -sS https://starship.rs/install.sh | sh -s -- -y

COPY --chown=${USERNAME}:${USERNAME} starship.toml /home/${USERNAME}/.config/starship.toml

RUN printf '#!/usr/bin/env zsh\neval "$(starship init zsh)"\n' > /home/${USERNAME}/.zshrc \
    && chown ${USERNAME}:${USERNAME} /home/${USERNAME}/.zshrc

ENV CLAUDE_CONFIG_DIR=/home/${USERNAME}/.claude
ENV GOPATH=/home/${USERNAME}/go
ENV GOROOT=/usr/local/go
ENV CGO_ENABLED=0
ENV SHELL=/bin/zsh
ENV EDITOR=vim
ENV VISUAL=vim

USER $USERNAME
