# config

A flexible configuration management package for Go applications that supports YAML, JSON, and environment variables.


## Installation

```sh
go get github.com/yc-alpha/config
```

## Usage

### Loading Configuration

```go
import "github.com/yc-alpha/config"

// Load from YAML file (default)
config.Load("config.yml")

// Load from JSON file
config.Load("config.json")
```

### Reading Values

```go
// Get with type conversion
port := config.GetInt("server.port", 8080)  // with default value
host := config.GetString("databases.main.host", "localhost")
debug := config.GetBool("logging.debug", false)

// Get nested values
dbVersion := config.Get("databases.main.version[0][1]")  // array access
phone := config.Get("address[0].phone")  // nested object with array

// Get raw value
value := config.Get("some.path")
```

### Setting Values

```go
// Set values with type safety
config.SetInt("server.port", 8080)
config.SetString("databases.main.host", "localhost")
config.SetBool("logging.debug", true)

// Set array values
config.SetString("versions[0]", "1.0.0")
config.SetString("databases.main.version[1][0]", "2.0.0")
```

### Environment Variables

Environment variables are automatically loaded and can override file-based configuration.

### Custom Configuration

```go
// Create custom configuration instance
cfg := config.New(
    config.WithSource(config.NewFile("config.yml")),
    config.WithSource(config.NewEnviron("APP_")), // With prefix
)

// Use custom instance
value := cfg.Get("some.path")
```


## License

MIT License - see [LICENSE](LICENSE) for more details.