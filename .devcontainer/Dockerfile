FROM golang:1.26.5-trixie

ARG TZ=UTC
ARG CLAUDE_CODE_VERSION=2.1.236

RUN apt-get update \
    && apt-get install -y --no-install-recommends \
       aggregate \
       ca-certificates \
       curl \
       dnsutils \
       fonts-firacode \
       fzf \
       gh \
       git \
       gnupg2 \
       iproute2 \
       ipset \
       iptables \
       jq \
       less \
       locales \
       man-db \
       nodejs \
       npm \
       procps \
       sudo \
       unzip \
       vim \
       wget \
       zsh \
    && rm -rf /var/lib/apt/lists/* \
    && sed -i 's/^# *\(en_US.UTF-8\)/\1/' /etc/locale.gen \
    && locale-gen en_US.UTF-8 \
    && update-locale LANG=en_US.UTF-8  \
    && apt-get autoremove -y \
    && apt-get clean -y

ENV TZ="${TZ}" \
    LANG="en_US.UTF-8" \
    LC_ALL="en_US.UTF-8" \
    DEVCONTAINER="true"

ARG USERNAME=sandbox
ARG USER_UID=1000
ARG USER_GID=${USER_UID}

RUN groupadd --gid "${USER_GID}" "${USERNAME}" \
    && useradd --uid "${USER_UID}" --gid "${USER_GID}" --create-home "${USERNAME}" \
    && chsh --shell /bin/zsh "${USERNAME}" \
    && mkdir -p  \
       /workspace \
       "/home/${USERNAME}/go/pkg/mod" \
       "/home/${USERNAME}/.cache/go-build" \
       "/home/${USERNAME}/.claude" \
       "/home/${USERNAME}/.config" \
       /usr/local/share/npm-global \
    && chown -R "${USERNAME}:${USERNAME}" \
       /workspace \
       /go "/home/${USERNAME}" \
       /usr/local/share/npm-global

WORKDIR /workspace

# Install Starship
COPY starship-install.sh /tmp/starship-install.sh

RUN sh /tmp/starship-install.sh -y \
    && rm -f /tmp/starship-install.sh

COPY --chown="${USERNAME}:${USERNAME}" \
     starship.toml \
     /home/${USERNAME}/.config/starship.toml

RUN printf '%s\n' \
           '#!/usr/bin/env zsh' \
           'eval "$(starship init zsh)"' \
    > "/home/${USERNAME}/.zshrc" \
    && chown "${USERNAME}:${USERNAME}" "/home/${USERNAME}/.zshrc"

# Firewall script
COPY init-firewall.sh /usr/local/bin/init-firewall.sh

RUN chmod 0755 /usr/local/bin/init-firewall.sh \
    && printf '%s\n' \
              "${USERNAME} ALL=(root) NOPASSWD: /usr/local/bin/init-firewall.sh" \
       > "/etc/sudoers.d/${USERNAME}-firewall" \
    && chmod 0440 "/etc/sudoers.d/${USERNAME}-firewall"

USER ${USERNAME}

ENV NPM_CONFIG_PREFIX="/usr/local/share/npm-global" \
    GOPATH="/home/${USERNAME}/go" \
    GOROOT="/usr/local/go" \
    CLAUDE_CONFIG_DIR="/home/${USERNAME}/.claude" \
    CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC="1" \
    DISABLE_AUTOUPDATER="1" \
    DO_NOT_TRACK="1" \
    SHELL="/bin/zsh" \
    EDITOR="vim" \
    VISUAL="vim" \
    PATH="${PATH}:/usr/local/share/npm-global/bin:${GOPATH}/bin:${GOROOT}/bin" \
    GOMODCACHE="/home/${USERNAME}/go/pkg/mod"

# Install Claude Code CLI
RUN npm install -g "@anthropic-ai/claude-code@${CLAUDE_CODE_VERSION}" \
    && printf '%s\n' '{"hasCompletedOnboarding": true}' \
       > "/home/${USERNAME}/.claude/.claude.json"
