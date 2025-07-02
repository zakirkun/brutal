# 🚀 Brutal Load Tester

A powerful, beautiful, and blazingly fast HTTP load testing tool with an interactive TUI (Text User Interface) and real-time analytics.

![Build Status](https://github.com/zakirkun/brutal/workflows/Continuous%20Integration/badge.svg)
![Release](https://github.com/zakirkun/brutal/workflows/Build%20and%20Release/badge.svg)
![Go Version](https://img.shields.io/badge/Go-1.23+-blue.svg)
![License](https://img.shields.io/badge/License-MIT-green.svg)

## ✨ Features

### 🎨 **Beautiful TUI Interface**
- **Real-time progress visualization** with animated progress bars
- **Live statistics** showing successful/failed requests, response times, and throughput
- **Interactive charts** with ASCII-based visualizations
- **Color-coded output** for easy status recognition
- **Responsive design** that adapts to terminal size

### 📊 **Advanced Analytics**
- **QPS Chart**: Real-time queries per second over the last 60 seconds
- **Response Time Chart**: Live response time visualization for the last 100 requests
- **Live Statistics**: Detailed metrics including RPS, avg/min/max response times
- **Data Transfer Tracking**: Monitor bandwidth usage in real-time
- **Status Code Analysis**: Comprehensive HTTP status code breakdown

### ⚡ **High Performance**
- **Concurrent request handling** with configurable worker pools
- **Efficient memory usage** with optimized data structures
- **Non-blocking UI updates** for smooth real-time experience
- **Lightweight binary** with minimal system requirements

### 🛠 **Flexible Configuration**
- **Multiple HTTP methods** (GET, POST, PUT, DELETE, etc.)
- **Custom headers** and request bodies
- **Configurable timeouts** and retry policies
- **TLS certificate verification control**
- **JSON output** for integration with other tools

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

## 🚀 Quick Start

### Basic Usage
```bash
# Simple load test
brutal https://api.example.com

# Custom configuration
brutal https://api.example.com -n 1000 -c 50 -timeout 10s
```

### Advanced Examples

#### POST Request with JSON Body
```bash
brutal https://api.example.com/users \
  -method POST \
  -headers '{"Content-Type": "application/json", "Authorization": "Bearer token"}' \
  -body '{"name": "John Doe", "email": "john@example.com"}'
```

#### High-Concurrency Test
```bash
brutal https://api.example.com \
  -n 10000 \
  -c 100 \
  -timeout 30s \
  -output results.json
```

#### Simple Mode (No TUI)
```bash
brutal https://api.example.com -no-tui -n 100 -c 10
```

## 📋 Usage

```
brutal [OPTIONS] <URL>

OPTIONS:
  -n int         Total number of requests (default: 100)
  -c int         Number of concurrent requests (default: 10)
  -timeout dur   Request timeout (default: 30s)
  -method string HTTP method (default: "GET")
  -headers string Headers in JSON format
  -body string   Request body
  -insecure      Skip TLS certificate verification
  -no-tui        Disable TUI and use simple output
  -output string Output file for JSON results
```

## 🎯 Example Output

### TUI Mode (Default)
```
🚀 Go Brutal Tester

┌─ Configuration ──────────────┐
│ URL: https://api.example.com │
│ Method: GET                  │
│ Concurrent: 50               │
│ Total Requests: 1000         │
│ Timeout: 30s                 │
└──────────────────────────────┘

Progress: 750/1000 (75.0%)
████████████████████████████████████████████████████████ 75.0%

┌─ Live Statistics ────────────┐
│ Elapsed: 15.2s               │
│ Successful: 745              │
│ Failed: 5                    │
│ Overall RPS: 49.34           │
│ Current QPS: 52.30           │
│ Avg Response Time: 145ms     │
│ Min Response Time: 89ms      │
│ Max Response Time: 287ms     │
│ Data Transferred: 2.15 MB    │
└──────────────────────────────┘

┌─ QPS Chart (last 60s) ───────┐
│ ████████████████████ 55.2    │
│ ██████████████████   48.7    │
│ ████████████████     42.1    │
│ ████████████         35.6    │
└──────────────────────────────┘

┌─ Response Time Chart ────────┐
│ ▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄ 250ms     │
│ ▄▄▄▄▄▄▄▄▄▄▄▄▄▄    200ms     │
│ ▄▄▄▄▄▄▄▄▄▄        150ms     │
└──────────────────────────────┘

⣻ Running... (Press Ctrl+C to stop)
```

### Simple Mode Output
```
Starting load test...
URL: https://api.example.com
Method: GET
Concurrent users: 10
Total requests: 100
--------------------------------------------------
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
brutal https://api.example.com -n 1000 -c 10 -timeout 5s
```

### API Endpoint Testing
```bash
# REST API
brutal https://api.example.com/v1/users -method GET
brutal https://api.example.com/v1/users -method POST -body '{"name":"test"}'

# GraphQL
brutal https://api.example.com/graphql \
  -method POST \
  -headers '{"Content-Type": "application/json"}' \
  -body '{"query": "{ users { id name } }"}'
```

## 📊 Output Formats

### JSON Output
Use `-output results.json` to save detailed results:
```json
{
  "total_requests": 1000,
  "successful_requests": 987,
  "failed_requests": 13,
  "total_time": "45.123s",
  "requests_per_second": 22.15,
  "data_transferred_mb": 5.67,
  "response_times": {
    "min": "89.123ms",
    "max": "1.234s", 
    "avg": "234.567ms",
    "p50": "220.123ms",
    "p95": "456.789ms",
    "p99": "987.654ms"
  },
  "status_codes": {
    "200": 987,
    "500": 13
  }
}
```

## 🛡️ Security Features

- **TLS Verification**: Enabled by default, can be disabled with `-insecure`
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
brutal https://api.example.com -timeout 60s
```

#### High Failure Rate
```bash
# Reduce concurrency
brutal https://api.example.com -c 5 -n 100

# Check server logs for rate limiting
brutal https://api.example.com -c 1 -n 10
```

#### TLS Certificate Issues
```bash
# Skip certificate verification (not recommended for production)
brutal https://api.example.com -insecure
```

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guidelines](CONTRIBUTING.md) for details.

### Development Setup
```bash
git clone https://github.com/zakirkun/brutal.git
cd brutal
go mod download
go run . -h
```

### Running Tests
```bash
go test ./...
go test -race ./...
go test -bench=. ./...
```

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Built with [Go](https://golang.org/)
- TUI powered by [Bubble Tea](https://github.com/charmbracelet/bubbletea)
- Inspired by tools like Apache Bench, wrk, and hey

---

**⭐ Star this project if you find it useful!**

**🐛 Found a bug? [Open an issue](https://github.com/zakirkun/brutal/issues)**

**💡 Have a feature request? [Start a discussion](https://github.com/zakirkun/brutal/discussions)** 