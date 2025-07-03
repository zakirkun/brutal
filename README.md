# 🚀 Brutal Load Tester

![Screenshot](./assets/screenshot.png)

A powerful, blazingly fast HTTP load testing tool with comprehensive analytics and a modern CLI interface powered by Cobra.

![Build Status](https://github.com/zakirkun/brutal/workflows/Continuous%20Integration/badge.svg)
![Release](https://github.com/zakirkun/brutal/workflows/Build%20and%20Release/badge.svg)
![Go Version](https://img.shields.io/badge/Go-1.23+-blue.svg)
![License](https://img.shields.io/badge/License-MIT-green.svg)

## ✨ Features

### ⚡ **High Performance**
- **Concurrent request handling** with configurable worker pools
- **Efficient memory usage** with optimized data structures
- **Real-time progress tracking** during load tests
- **Lightweight binary** with minimal system requirements

### 📊 **Comprehensive Analytics**
- **Detailed statistics** including response times, throughput, and error rates
- **Percentile analysis** (50th, 95th, 99th percentiles)
- **Status code breakdown** with detailed error reporting
- **Data transfer tracking** with bandwidth monitoring
- **JSON output** for integration with other tools

### 🛠 **Modern CLI Interface**
- **Cobra-powered CLI** with intuitive commands and flags
- **Shell completion** support (Bash, Zsh, Fish, PowerShell)
- **Contextual help** with examples and usage patterns
- **Version management** with dedicated commands
- **Flexible input** - URL as argument or flag
- **ASCII art banner** with "Brutal Buster" branding (optional)

### 🔧 **Flexible Configuration**
- **Multiple HTTP methods** (GET, POST, PUT, DELETE, etc.)
- **Custom headers** and request bodies
- **Configurable timeouts** and retry policies
- **TLS certificate verification control**
- **Short and long flag options** for convenience

## 📦 Installation

### Quick Install (Recommended)

#### Windows
```bash
# Download and run installer
curl -LO https://github.com/zakirkun/brutal/releases/latest/download/brutal-windows-installer.zip
unzip brutal-windows-installer.zip
# Run install.bat as Administrator
```

#### Linux (Ubuntu/Debian)
```bash
# Install DEB package
wget https://github.com/zakirkun/brutal/releases/latest/download/brutal-1.0.1-amd64.deb
sudo dpkg -i brutal-1.0.1-amd64.deb
```

#### Linux (CentOS/RHEL/Fedora)
```bash
# Install RPM package
wget https://github.com/zakirkun/brutal/releases/latest/download/brutal-1.0.1-1.x86_64.rpm
sudo rpm -i brutal-1.0.1-1.x86_64.rpm
```

#### macOS
```bash
# Download and install universal binary
curl -LO https://github.com/zakirkun/brutal/releases/latest/download/brutal-macos-installer.tar.gz
tar -xzf brutal-macos-installer.tar.gz
./install.sh
```

### Manual Installation

Download the appropriate binary for your platform from the [releases page](https://github.com/zakirkun/brutal/releases):

- **Windows**: `brutal-windows-amd64.exe` or `brutal-windows-arm64.exe`
- **Linux**: `brutal-linux-amd64` or `brutal-linux-arm64`  
- **macOS**: `brutal-darwin-amd64` or `brutal-darwin-arm64`

Place the binary in your PATH and make it executable (Unix systems):
```bash
chmod +x brutal-*
sudo mv brutal-* /usr/local/bin/brutal
```

### Build from Source
```bash
git clone https://github.com/zakirkun/brutal.git
cd brutal
go build -o brutal .
```

### Shell Completion Setup

After installation, enable shell completion for a better experience:

#### Bash
```bash
# Add to ~/.bashrc
echo 'source <(brutal completion bash)' >> ~/.bashrc

# Or install system-wide (Linux)
sudo brutal completion bash > /etc/bash_completion.d/brutal
```

#### Zsh
```bash
# Add to ~/.zshrc
echo 'autoload -U compinit; compinit' >> ~/.zshrc
brutal completion zsh > "${fpath[1]}/_brutal"
```

#### Fish
```bash
brutal completion fish > ~/.config/fish/completions/brutal.fish
```

#### PowerShell
```powershell
brutal completion powershell | Out-String | Invoke-Expression
```

## 🚀 Quick Start

### Basic Usage
```bash
# Simple load test (URL as argument)
brutal https://api.example.com

# Using URL flag
brutal --url https://api.example.com

# Custom configuration with short flags
brutal https://api.example.com -n 1000 -c 50 -t 10s

# Custom configuration with long flags
brutal https://api.example.com --requests 1000 --concurrent 50 --timeout 10s

# With proxy support
brutal https://api.example.com -n 100 -c 10 --proxy http://proxy.example.com:8080
```

### Advanced Examples

#### POST Request with JSON Body
```bash
# Using short flags
brutal https://api.example.com/users \
  -X POST \
  -H '{"Content-Type": "application/json", "Authorization": "Bearer token"}' \
  -d '{"name": "John Doe", "email": "john@example.com"}'

# Using long flags
brutal https://api.example.com/users \
  --method POST \
  --headers '{"Content-Type": "application/json", "Authorization": "Bearer token"}' \
  --body '{"name": "John Doe", "email": "john@example.com"}'
```

#### High-Concurrency Test with Output
```bash
brutal https://api.example.com \
  --requests 10000 \
  --concurrent 100 \
  --timeout 30s \
  --output results.json
```

#### Proxy Testing
```bash
# HTTP proxy
brutal https://api.example.com --proxy http://proxy.example.com:8080

# SOCKS5 proxy
brutal https://api.example.com --proxy socks5://proxy.example.com:1080

# Authenticated proxy
brutal https://api.example.com --proxy http://username:password@proxy.example.com:8080

# Test through corporate proxy
brutal https://httpbin.org/get \
  --proxy http://corporate-proxy:3128 \
  --requests 10 \
  --concurrent 2
```

#### Insecure HTTPS Testing
```bash
brutal https://self-signed-cert.example.com \
  --insecure \
  --requests 50 \
  --concurrent 5
```

## 📋 Command Reference

### Main Command
```bash
brutal [URL] [flags]
```

### Subcommands
```bash
brutal version                    # Show version information
brutal completion [shell]         # Generate shell completion scripts
brutal help                      # Show help for any command
```

### Flags

| Short | Long          | Default | Description                           |
|-------|---------------|---------|---------------------------------------|
| `-u`  | `--url`       | -       | Target URL to test                    |
| `-X`  | `--method`    | GET     | HTTP method                           |
| `-H`  | `--headers`   | -       | Headers in JSON format                |
| `-d`  | `--body`      | -       | Request body                          |
| `-c`  | `--concurrent`| 10      | Number of concurrent requests         |
| `-n`  | `--requests`  | 100     | Total number of requests              |
| `-t`  | `--timeout`   | 30s     | Request timeout                       |
| `-k`  | `--insecure`  | false   | Skip TLS certificate verification     |
| `-o`  | `--output`    | -       | Output file for JSON results          |
| `-p`  | `--proxy`     | -       | Proxy URL (http/https/socks5)         |
|       | `--no-banner` | false   | Disable ASCII art banner              |
| `-h`  | `--help`      | -       | Help for brutal                       |

### Examples with Different Flag Styles
```bash
# Short flags (compact)
brutal https://api.example.com -n 100 -c 10 -t 5s -o results.json

# Long flags (readable)
brutal https://api.example.com --requests 100 --concurrent 10 --timeout 5s --output results.json

# Mixed flags
brutal https://api.example.com -n 100 --concurrent 10 -t 5s --output results.json

# URL as flag vs argument
brutal --url https://api.example.com -n 100
brutal https://api.example.com -n 100

# With proxy
brutal https://api.example.com -n 100 --proxy http://proxy:8080

# Disable banner for scripts/CI/CD
brutal https://api.example.com -n 100 --no-banner
```

## 🎯 Example Output

```
Starting load test...
URL: https://api.example.com
Method: GET
Concurrent users: 10
Total requests: 100
Timeout: 30s
--------------------------------------------------
Progress: 100/100 (100.0%)
Completed: 100/100 (100.0%)
============================================================
LOAD TEST RESULTS
============================================================
Total Requests: 100
Successful: 98 (98.00%)
Failed: 2 (2.00%)
Total Time: 5.234s
Requests/sec: 19.11
Data Transfer: 0.85 MB
----------------------------------------
RESPONSE TIMES
----------------------------------------
Min: 89.245ms
Max: 456.123ms
Avg: 187.456ms
50th percentile: 165.234ms
95th percentile: 398.567ms
99th percentile: 445.123ms
----------------------------------------
STATUS CODES
----------------------------------------
200: 98 (98.0%)
500: 2 (2.0%)
============================================================
```

## 🔧 Configuration Examples

### Environment-Specific Testing
```bash
# Development
brutal https://dev-api.example.com -n 100 -c 5

# Staging  
brutal https://staging-api.example.com -n 500 -c 20

# Production (careful!)
brutal https://api.example.com -n 1000 -c 10 -t 5s
```

### API Endpoint Testing
```bash
# REST API
brutal https://api.example.com/v1/users -X GET
brutal https://api.example.com/v1/users -X POST -d '{"name":"test"}'

# GraphQL
brutal https://api.example.com/graphql \
  --method POST \
  --headers '{"Content-Type": "application/json"}' \
  --body '{"query": "{ users { id name } }"}'
```

### Different Authentication Methods
```bash
# Bearer Token
brutal https://api.example.com \
  -H '{"Authorization": "Bearer your-token-here"}' \
  -n 50

# API Key
brutal https://api.example.com \
  -H '{"X-API-Key": "your-api-key"}' \
  -n 50

# Basic Auth (Base64 encoded)
brutal https://api.example.com \
  -H '{"Authorization": "Basic dXNlcjpwYXNz"}' \
  -n 50
```

### Banner Control
```bash
# With banner (default) - great for interactive use
brutal https://api.example.com -n 100

# Without banner - perfect for scripts and CI/CD
brutal https://api.example.com -n 100 --no-banner

# Version with banner
brutal version

# Version without banner for scripts
brutal version --no-banner
```

## 📊 Output Formats

### Console Output
Beautiful, real-time console output with progress indicators and comprehensive statistics.

### JSON Output
Use `-o results.json` or `--output results.json` to save detailed results:
```json
{
  "config": {
    "url": "https://api.example.com",
    "method": "GET",
    "concurrent": 10,
    "requests": 100,
    "timeout": "30s"
  },
  "stats": {
    "total_requests": 100,
    "successful_requests": 98,
    "failed_requests": 2,
    "total_time": "5.234s",
    "requests_per_second": 19.11,
    "data_transferred_mb": 0.85,
    "response_times": {
      "min": "89.123ms",
      "max": "456.123ms", 
      "avg": "187.456ms",
      "p50": "165.234ms",
      "p95": "398.567ms",
      "p99": "445.123ms"
    },
    "status_codes": {
      "200": 98,
      "500": 2
    }
  }
}
```

## 🛡️ Security Features

- **TLS Verification**: Enabled by default, can be disabled with `-k`/`--insecure`
- **Safe Defaults**: Conservative default values to prevent accidental DoS
- **No Sensitive Data Logging**: Ensures credentials aren't leaked in outputs
- **Timeout Protection**: Prevents hanging requests

## 🔍 Troubleshooting

### Common Issues

#### Connection Refused
```bash
# Check if the target server is running
curl -I https://api.example.com

# Test with a longer timeout
brutal https://api.example.com --timeout 60s
```

#### High Failure Rate
```bash
# Reduce concurrency
brutal https://api.example.com --concurrent 5 --requests 100

# Check server logs for rate limiting
brutal https://api.example.com -c 1 -n 10
```

#### TLS Certificate Issues
```bash
# Skip certificate verification (not recommended for production)
brutal https://api.example.com --insecure
```

#### Help and Documentation
```bash
# General help
brutal --help

# Command-specific help
brutal completion --help

# Version information
brutal version
```

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guidelines](CONTRIBUTING.md) for details.

### Development Setup
```bash
git clone https://github.com/zakirkun/brutal.git
cd brutal
go mod download
go run . --help
```

### Running Tests
```bash
go test ./...
go test -race ./...
go test -bench=. ./...
```

### Testing CLI Changes
```bash
# Test help system
go run . --help
go run . version
go run . completion bash

# Test load testing
go run . https://httpbin.org/get -n 5 -c 2
```

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Built with [Go](https://golang.org/)
- CLI powered by [Cobra](https://github.com/spf13/cobra)
- Inspired by tools like Apache Bench, wrk, and hey

---

**⭐ Star this project if you find it useful!**

**🐛 Found a bug? [Open an issue](https://github.com/zakirkun/brutal/issues)**

**💡 Have a feature request? [Start a discussion](https://github.com/zakirkun/brutal/discussions)**

### Network and Proxy Testing
```bash
# Test through different proxy types
brutal https://httpbin.org/get --proxy http://proxy.example.com:8080
brutal https://httpbin.org/get --proxy socks5://proxy.example.com:1080

# Corporate network testing
brutal https://internal-api.company.com \
  --proxy http://corporate-proxy:3128 \
  --insecure \
  --requests 50

# Compare direct vs proxy performance
brutal https://api.example.com -n 100 -c 10
brutal https://api.example.com -n 100 -c 10 --proxy http://proxy:8080
``` 