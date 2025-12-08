# Node.js sandbox for code execution
FROM node:20-alpine

# Install common Node.js packages globally
RUN npm install -g \
    typescript \
    ts-node \
    @types/node

# Create non-root user
RUN addgroup -S sandbox && adduser -S sandbox -G sandbox

# Create working directory
WORKDIR /workspace
RUN chown sandbox:sandbox /workspace

# Switch to non-root user
USER sandbox

# Default command
CMD ["node"]
