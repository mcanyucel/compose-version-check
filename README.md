# Docker Compose Checker

A Go application that monitors Docker Compose files for changes by comparing local files with their upstream sources. It can notify you of any image version changes via Slack, ntfy.sh, or Telegram.

## Features

- 🔍 Monitors multiple Docker Compose files concurrently
- 🔄 Compares local files with upstream sources
- 🎯 Detects changes in image versions
- 📧 Notifications through:
  - Slack (via webhooks)
  - ntfy.sh
  - Telegram
  - Debug output (console/file)
- ⚡ Parallel processing of multiple files
- 🔒 Support for both public and private repositories

## System Requirements

- x86_64 (64-bit) architecture
- Linux/Unix-like operating system

Note: For other architectures (like ARM), you'll need to build from source.

## Installation

Choose one of these methods to install compose-checker:

### Quick Install Script (Linux/macOS, x86_64 only)

```bash
curl -sSL https://raw.githubusercontent.com/mcanyucel/compose-version-check/main/install.sh | bash
```

This will:
- Check if your system is compatible (x86_64)
- Place the binary in ~/.local/bin (or /usr/local/bin if run as root)
- Create a template config.yaml in ~/.config/compose-checker/

### Manual Installation

1. Download the latest release from [GitHub Releases](https://github.com/mcanyucel/compose-version-check/releases)
2. Extract and move the binary:
```bash
# Linux/macOS (x86_64)
chmod +x compose-checker
sudo mv compose-checker /usr/local/bin/

# Or without sudo to your user bin directory:
mkdir -p ~/.local/bin
mv compose-checker ~/.local/bin/
```

3. Create a config file:
```bash
mkdir -p ~/.config/compose-checker
curl -sSL https://raw.githubusercontent.com/mcanyucel/compose-version-check/main/config.yaml.example > ~/.config/compose-checker/config.yaml
```

### Building from Source (for other architectures)

If you're not on x86_64 architecture, you'll need to build from source:

```bash
git clone https://github.com/mcanyucel/compose-version-check.git
cd compose-version-check
go build -o compose-checker
```

### Docker

You can run compose-checker in a container with automated scheduling using Ofelia. Note that the Docker image is built for x86_64 architecture.

The container runs an initial check immediately upon startup and then runs periodic checks based on the configured interval.

```bash
docker run -d \
  --name compose-checker \
  -v /path/to/your/compose/files:/watch:ro \
  -v /path/to/config.yaml:/app/config.yaml:ro \
  -e CHECK_INTERVAL=6 \  # Optional: Check every 6 hours (default)
  mcanyucel/compose-checker
```

Example config file (works for both Docker and non-Docker usage):
```yaml
files:
  - local_path: "docker-compose.yaml"
    source_url: "https://raw.githubusercontent.com/user/repo/main/docker-compose.yaml"
  - local_path: "other/docker-compose.yaml"
    source_url: "https://raw.githubusercontent.com/user/other/main/docker-compose.yaml"
notifications:
  type: "slack"  # or "ntfy" or "telegram" or "debug"
  slack_webhook: "https://hooks.slack.com/services/YOUR/WEBHOOK/URL"
```

Docker Compose example:
```yaml
version: '3'
services:
  compose-checker:
    image: mcanyucel/compose-checker
    volumes:
      - /path/to/your/compose/files:/watch:ro
      - ./config.yaml:/app/config.yaml:ro
    environment:
      - CHECK_INTERVAL=6  # Check every 6 hours (default)
    restart: unless-stopped
```

Note: When running in Docker:
- The application automatically detects it's in a container and handles path mapping
- Uses Ofelia for reliable container-native scheduling
- Runs an initial check immediately upon startup
- Performs subsequent checks every CHECK_INTERVAL hours (defaults to 6 if not specified)
- You don't need to modify paths in the config file - just use the paths as they appear in your filesystem
- All compose files should be within the mounted directory


## Configuration

Create a `config.yaml` file:

```yaml
files:
  - local_path: "./docker-compose.yaml"
    source_url: "https://raw.githubusercontent.com/user/repo/main/docker-compose.yaml"
  - local_path: "./project2/docker-compose.yaml"
    source_url: "https://raw.githubusercontent.com/user/project2/main/docker-compose.yaml"
notifications:
  type: "slack"  # or "ntfy" or "telegram" or "debug"
  slack_webhook: "https://hooks.slack.com/services/YOUR/WEBHOOK/URL"  # for Slack
  ntfy_topic: "your-topic"  # for ntfy
  ntfy_server: "https://ntfy.sh"  # optional for ntfy, defaults to https://ntfy.sh
  telegram_token: "your-bot-token"  # for Telegram
  telegram_chat: "your-chat-id"  # for Telegram
  debug_file: "notifications"  # optional, for debug mode
```
### Configuration File Generator

The python script file *compose_finder.py* can traverse the given root directory to find *docker-compose.yaml* files (recursively), copy them to this project root, into *containers* directory, and generate the skeleton of the configuration file. It requires pyyaml, so the suggested method of execution is to use the *compose-finder.sh* bash script, which creates a new python environment (if not exists), activating it, installing necessary packages, then running the script (finally deactivating the environment). This script also accepts the search root folder as an argument, and passes it to the python script.

Before first use, don't forget to make the sh file executable:

```
chmod +x compose-finder.sh
```

Then you can run the script:

```
./compose-finder.sh /path/to/source/directory
```

## Usage

### Basic Usage

Run the checker with default config:
```bash
./compose-checker
```

Any changes will be reported to your configured notification service:
[Slack](https://api.slack.com/messaging/webhooks), [ntfy.sh](https://ntfy.sh), or Telegram.

![image](https://imgurl.mustafacanyucel.com/i/52d94f23-e86e-4986-b950-6f8963e093a0.jpg)

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

### Using Cron (Non-Docker)

Add to crontab to run every 6 hours:
```bash
0 */6 * * * /path/to/compose-checker -config /path/to/config.yaml
```

### Using Docker

When using Docker, scheduling is handled automatically by Ofelia within the container. You can configure the check interval using the CHECK_INTERVAL environment variable (either in compose file or as below):

```bash
# Check every 2 hours
CHECK_INTERVAL=2 docker compose up -d

# Check every 12 hours
CHECK_INTERVAL=12 docker compose up -d

# Use default 6-hour interval
docker compose up -d
```

## Notification Examples

### Slack
Messages will be formatted and sent to your configured Slack webhook URL.

### ntfy.sh
Notifications will be sent to your configured ntfy.sh topic.

### Telegram
To use Telegram notifications:

1. Create a new bot:
   - Message [@BotFather](https://t.me/botfather) on Telegram
   - Use the `/newbot` command and follow the instructions
   - Save the bot token you receive

2. Get your chat ID:
   - Send a message to your new bot
   - Visit `https://api.telegram.org/bot<YourBotToken>/getUpdates`
   - Look for the `chat.id` field in the response

3. Configure in your `config.yaml`:
```yaml
notifications:
  type: telegram
  telegram_token: "your-bot-token"
  telegram_chat: "your-chat-id"
```

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
