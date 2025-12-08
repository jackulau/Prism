# Shell sandbox for command execution
FROM alpine:3.18

# Install common shell utilities
RUN apk add --no-cache \
    bash \
    coreutils \
    findutils \
    grep \
    sed \
    awk \
    curl \
    wget \
    git \
    jq \
    yq \
    openssh-client \
    && rm -rf /var/cache/apk/*

# Create non-root user
RUN addgroup -S sandbox && adduser -S sandbox -G sandbox

# Create working directory
WORKDIR /workspace
RUN chown sandbox:sandbox /workspace

# Switch to non-root user
USER sandbox

# Default command
CMD ["/bin/bash"]
