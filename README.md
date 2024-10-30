# Docker Compose Checker

A Go application that monitors Docker Compose files for changes by comparing local files with their upstream sources. It can notify you of any image version changes via Slack or ntfy.sh.

## Features

- 🔍 Monitors multiple Docker Compose files concurrently
- 🔄 Compares local files with upstream sources
- 🎯 Detects changes in image versions
- 📧 Notifications through:
  - Slack (via webhooks)
  - ntfy.sh
  - Debug output (console/file)
- ⚡ Parallel processing of multiple files
- 🔒 Support for both public and private repositories

## Installation

1. Clone the repository:
```bash
git clone [your-repo-url]
cd docker-compose-checker
```

2. Build the application:
```bash
go build -o compose-checker
```

3. Make it executable:
```bash
chmod +x compose-checker
```

## Configuration

Create a `config.yaml` file:

```yaml
files:
  - local_path: "./docker-compose.yaml"
    source_url: "https://raw.githubusercontent.com/user/repo/main/docker-compose.yaml"
  - local_path: "./project2/docker-compose.yaml"
    source_url: "https://raw.githubusercontent.com/user/project2/main/docker-compose.yaml"
notifications:
  type: "slack"  # or "ntfy" or "debug"
  slack_webhook: "https://hooks.slack.com/services/YOUR/WEBHOOK/URL"  # for Slack
  ntfy_topic: "your-topic"  # for ntfy
  ntfy_server: "https://ntfy.sh"  # optional for ntfy, defaults to https://ntfy.sh
  debug_file: "notifications"  # optional, for debug mode
```

## Usage

### Basic Usage

Run the checker with default config:
```bash
./compose-checker
```

### Debug Mode

Run with debug output:
```bash
./compose-checker -debug
```

### Custom Config

Use a specific config file:
```bash
./compose-checker -config path/to/config.yaml
```

## Automated Checking

### Using Cron

Add to crontab to run every 6 hours:
```bash
0 */6 * * * /path/to/compose-checker -config /path/to/config.yaml
```

### Using a Shell Script

Create a shell script for automated runs:

```bash
#!/bin/bash
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
$DIR/compose-checker -config $DIR/config.yaml
```

## Notification Examples

### Slack
Messages will be formatted and sent to your configured Slack webhook URL.

### ntfy.sh
Notifications will be sent to your configured ntfy.sh topic.

### Debug Output
When using debug mode, output will look like:
```
=== Debug Notification [2024-01-01T12:00:00Z] ===
Changes found in ./docker-compose.yaml:
  Service web:
    Old image: nginx:1.19
    New image: nginx:1.20
===============================
```

## Development

### Prerequisites

- Go 1.16 or later

### Setup Development Environment

1. Install dependencies:
```bash
go mod download
```

2. Create test files:
```bash
mkdir -p test
cat > test/docker-compose.yaml << EOF
services:
  web:
    image: nginx:1.19
  db:
    image: postgres:13
EOF
```

3. Create debug config:
```bash
cp config.yaml.example config.yaml
```

### Build and Test

Build the application:
```bash
go build
```

Run with debug mode:
```bash
./compose-checker -debug
```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgments

- Built with Go's concurrency features
- Uses yaml.v3 for YAML parsing
- Inspired by the need for Docker Compose version tracking