# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Initial release of yoyo-nodered-wrapper
- Core wrapper functionality for Node-RED integration
- Flow deployment, execution, and management
- Custom converter and executor interfaces
- Comprehensive examples and CLI tools
- Yoyo-specific integration examples
- Full test coverage
- Documentation and build scripts

### Features
- **Flow Management**: Deploy, execute, and manage Node-RED flows
- **Custom Converters**: Convert between different workflow formats
- **Execution Handlers**: Pre/post execution hooks and error handling
- **Health Checking**: Monitor Node-RED instance health
- **CLI Tool**: Command-line interface for flow operations
- **Examples**: Basic, advanced, and Yoyo integration examples
- **Testing**: Comprehensive test suite with mocking support

### Dependencies
- Go 1.21+
- Node-RED (for actual flow execution)
- testify (for testing)
- yaml.v3 (for configuration)
