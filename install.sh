#!/bin/bash

# Detect architecture
ARCH=$(uname -m)

# Check if architecture is supported
if [ "$ARCH" != "x86_64" ]; then
    echo "Error: This binary is only available for x86_64 architecture."
    echo "Your architecture is: $ARCH"
    echo "Please build from source for your architecture."
    exit 1
fi

# Set installation directory
if [ "$EUID" -eq 0 ]; then
    INSTALL_DIR="/usr/local/bin"
    CONFIG_DIR="/etc/compose-checker"
else
    INSTALL_DIR="$HOME/.local/bin"
    CONFIG_DIR="$HOME/.config/compose-checker"
fi

# Create directories
mkdir -p "$INSTALL_DIR"
mkdir -p "$CONFIG_DIR"

# Download latest release
echo "Downloading latest release..."
RELEASE_URL="https://github.com/mcanyucel/compose-version-check/releases/latest/download/compose-checker"
curl -sSL "$RELEASE_URL" -o "$INSTALL_DIR/compose-checker"
chmod +x "$INSTALL_DIR/compose-checker"

# Download example config if none exists
if [ ! -f "$CONFIG_DIR/config.yaml" ]; then
    echo "Downloading example config..."
    curl -sSL "https://raw.githubusercontent.com/mcanyucel/compose-version-check/main/config.yaml.example" -o "$CONFIG_DIR/config.yaml"
fi

echo "Installation complete!"
echo "Binary installed to: $INSTALL_DIR/compose-checker"
echo "Config file location: $CONFIG_DIR/config.yaml"
echo "Please edit the config file before running compose-checker"