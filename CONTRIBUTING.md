# Contributing to Yoyo Node-RED Wrapper

Thank you for your interest in contributing to the Yoyo Node-RED Wrapper! This document provides guidelines and information for contributors.

## Getting Started

### Prerequisites

- Go 1.21 or later
- Node-RED (for testing)
- Git

### Development Setup

1. **Fork and clone the repository**
   ```bash
   git clone https://github.com/yourusername/yoyo-nodered-wrapper.git
   cd yoyo-nodered-wrapper
   ```

2. **Install dependencies**
   ```bash
   go mod tidy
   go mod download
   ```

3. **Install development tools**
   ```bash
   make dev-setup
   ```

4. **Run tests**
   ```bash
   make test
   ```

## Development Workflow

### Code Style

- Follow Go's standard formatting (`gofmt`)
- Use `golangci-lint` for linting
- Write comprehensive tests for new features
- Document public APIs with Go doc comments

### Testing

- Write unit tests for all new functionality
- Ensure existing tests continue to pass
- Add integration tests for complex features
- Test with different Node-RED versions when possible

### Pull Request Process

1. **Create a feature branch**
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes**
   - Write code following the style guidelines
   - Add tests for new functionality
   - Update documentation as needed

3. **Test your changes**
   ```bash
   make test
   make lint
   ```

4. **Commit your changes**
   ```bash
   git add .
   git commit -m "feat: add your feature description"
   ```

5. **Push and create a pull request**
   ```bash
   git push origin feature/your-feature-name
   ```

### Commit Message Format

We follow the [Conventional Commits](https://www.conventionalcommits.org/) specification:

- `feat:` for new features
- `fix:` for bug fixes
- `docs:` for documentation changes
- `style:` for formatting changes
- `refactor:` for code refactoring
- `test:` for test additions/changes
- `chore:` for maintenance tasks

## Project Structure

```
yoyo-nodered-wrapper/
├── cmd/                    # CLI tools and examples
├── internal/              # Internal packages
│   ├── client/           # Node-RED HTTP client
│   ├── config/           # Configuration management
│   └── monitoring/       # Health checks and metrics
├── pkg/                   # Public packages
│   ├── wrapper/          # Main wrapper package
│   ├── converter/        # Workflow converters
│   ├── executor/         # Execution handlers
│   └── types/            # Common types
├── examples/              # Usage examples
├── docs/                  # Documentation
├── scripts/               # Build and utility scripts
└── test/                  # Test utilities and data
```

## Adding New Features

### Custom Converters

To add a new workflow converter:

1. Implement the `WorkflowConverter` interface
2. Add tests for your converter
3. Update documentation with usage examples

### Custom Executors

To add a new execution handler:

1. Implement the `ExecutionHandler` interface
2. Add tests for your executor
3. Document the behavior and configuration options

### New Node Types

To add support for new Node-RED node types:

1. Update the `Node` type if needed
2. Add conversion logic in converters
3. Add tests for the new node type
4. Update documentation

## Reporting Issues

When reporting issues, please include:

- Go version
- Node-RED version
- Operating system
- Steps to reproduce
- Expected vs actual behavior
- Error messages and logs

## Code of Conduct

This project follows the [Contributor Covenant](https://www.contributor-covenant.org/) Code of Conduct.

## License

By contributing to this project, you agree that your contributions will be licensed under the same license as the project (MIT License).

## Questions?

If you have questions about contributing, please:

- Open an issue for discussion
- Check existing issues and pull requests
- Review the documentation in the `docs/` directory
