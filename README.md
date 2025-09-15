# Yoyo Node-RED Wrapper

A high-level Go wrapper for managing Node-RED workflows. This library provides a clean, idiomatic Go interface for deploying, executing, and managing Node-RED flows.

## Features

- 🚀 **Easy Deployment**: Deploy workflows to Node-RED with a simple API
- ⚡ **Flow Execution**: Trigger and monitor workflow executions
- 🔧 **Configurable**: Flexible configuration options for different environments
- 📊 **Monitoring**: Built-in metrics and health checking
- 🧪 **Testable**: Comprehensive test utilities and mocking support
- 📚 **Well Documented**: Extensive documentation and examples

## Installation

```bash
go get github.com/yourusername/yoyo-nodered-wrapper
```

## Quick Start

```go
package main

import (
    "context"
    "log"
    
    nodered "github.com/yoyo-mq/go-nodered-wrapper/pkg/wrapper"
)

func main() {
    // Create configuration
    config := &nodered.Config{
        NodeRedURL: "http://localhost:1880",
        APIKey:     "your-api-key",
        Timeout:    30 * time.Second,
        Debug:      true,
    }
    
    // Create wrapper instance
    wrapper, err := nodered.New(config)
    if err != nil {
        log.Fatal(err)
    }
    
    // Deploy a workflow
    flow := &nodered.FlowDefinition{
        ID:   "my-workflow",
        Name: "My Workflow",
        // ... define your flow
    }
    
    err = wrapper.DeployFlow(context.Background(), flow)
    if err != nil {
        log.Fatal(err)
    }
    
    // Execute the workflow
    result, err := wrapper.ExecuteFlow(context.Background(), "my-workflow", map[string]interface{}{
        "input": "data",
    })
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("Execution result: %+v", result)
}
```

## Examples

- [Basic Usage](examples/basic/)
- [Advanced Configuration](examples/advanced/)
- [Yoyo Integration](examples/yoyo-integration/)

## Documentation

- [API Reference](docs/api.md)
- [Configuration Guide](docs/configuration.md)
- [Custom Nodes](docs/custom-nodes.md)
- [Contributing](CONTRIBUTING.md)

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Contributing

Contributions are welcome! Please read our [Contributing Guide](CONTRIBUTING.md) for details on our code of conduct and the process for submitting pull requests.

## Changelog

See [CHANGELOG.md](CHANGELOG.md) for a list of changes and version history.
