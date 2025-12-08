# Contributing to Prism

Thank you for your interest in contributing to Prism! This document provides guidelines and instructions for contributing.

## Getting Started

1. Fork the repository
2. Clone your fork locally
3. Set up the development environment (see README.md)
4. Create a new branch for your feature or fix

## Development Setup

```bash
# Install dependencies
make setup

# Start development servers
make dev
```

## Pull Request Process

1. Ensure your code follows the existing style and conventions
2. Update documentation if you're changing functionality
3. Add tests for new features
4. Ensure all tests pass before submitting
5. Create a pull request with a clear description of the changes

## Code Style

### Backend (Go)
- Follow standard Go formatting (`go fmt`)
- Use meaningful variable and function names
- Add comments for exported functions

### Frontend (TypeScript/React)
- Use TypeScript for type safety
- Follow the existing component structure
- Use functional components with hooks

## Reporting Issues

When reporting issues, please include:
- A clear description of the problem
- Steps to reproduce
- Expected vs actual behavior
- Environment details (OS, browser, versions)

## Security

If you discover a security vulnerability, please report it privately rather than opening a public issue. See the README for contact information.

## License

By contributing, you agree that your contributions will be licensed under the Apache License 2.0.
