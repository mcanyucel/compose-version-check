#!/bin/bash

# Directory where the script is located
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
VENV_DIR="$SCRIPT_DIR/.venv"
REQUIREMENTS_FILE="$SCRIPT_DIR/requirements.txt"
PYTHON_SCRIPT="$SCRIPT_DIR/compose_finder.py"

# Create requirements.txt if it doesn't exist
if [ ! -f "$REQUIREMENTS_FILE" ]; then
    echo "pyyaml" > "$REQUIREMENTS_FILE"
fi

# Check if Python 3 is installed
if ! command -v python3 &> /dev/null; then
    echo "Error: Python 3 is required but not installed."
    echo "Please install Python 3 using your package manager:"
    echo "  Debian/Ubuntu: sudo apt-get install python3 python3-venv"
    echo "  RHEL/CentOS: sudo dnf install python3"
    exit 1
fi

# Create virtual environment if it doesn't exist
if [ ! -d "$VENV_DIR" ]; then
    echo "Creating Python virtual environment..."
    python3 -m venv "$VENV_DIR"
    if [ $? -ne 0 ]; then
        echo "Error: Failed to create virtual environment."
        echo "Please make sure python3-venv is installed:"
        echo "  Debian/Ubuntu: sudo apt-get install python3-venv"
        exit 1
    fi
fi

# Activate virtual environment
source "$VENV_DIR/bin/activate"

# Install/upgrade pip and requirements
echo "Installing/updating required packages..."
python3 -m pip install --upgrade pip > /dev/null
python3 -m pip install -r "$REQUIREMENTS_FILE" > /dev/null

# Check if we received any arguments
if [ $# -eq 0 ]; then
    echo "Usage: $0 <source_directory>"
    exit 1
fi

# Run the Python script with all arguments passed to this script
python3 "$PYTHON_SCRIPT" "$@"

# Deactivate virtual environment
deactivate
