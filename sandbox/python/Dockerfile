# Python sandbox for code execution
FROM python:3.11-alpine

# Install common Python packages with pinned versions for reproducible builds
# Update these versions periodically after security review
RUN pip install --no-cache-dir \
    requests==2.31.0 \
    beautifulsoup4==4.12.2 \
    pandas==2.1.4 \
    numpy==1.26.2 \
    matplotlib==3.8.2 \
    pillow==10.1.0 \
    pyyaml==6.0.1 \
    httpx==0.25.2 \
    aiohttp==3.9.1

# Create non-root user
RUN addgroup -S sandbox && adduser -S sandbox -G sandbox

# Create working directory
WORKDIR /workspace
RUN chown sandbox:sandbox /workspace

# Switch to non-root user
USER sandbox

# Default command
CMD ["python3"]
