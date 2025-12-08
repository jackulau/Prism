# Base sandbox image for secure code execution
FROM alpine:3.18

# Install common utilities
RUN apk add --no-cache \
    bash \
    curl \
    wget \
    git \
    jq \
    && rm -rf /var/cache/apk/*

# Create non-root user for execution
RUN addgroup -S sandbox && adduser -S sandbox -G sandbox

# Create working directory
WORKDIR /workspace
RUN chown sandbox:sandbox /workspace

# Set resource limits via ulimit
# These are additional limits on top of Docker's cgroup limits

# Switch to non-root user
USER sandbox

# Default command
CMD ["/bin/bash"]
