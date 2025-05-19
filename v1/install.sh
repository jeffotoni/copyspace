#!/bin/bash
# Copyspace Installer - by jeffotoni

set -e

EXEC="copyspace"
INSTALL_DIR="/usr/local/bin"
BIN_PATH="$INSTALL_DIR/$EXEC"
DOKEYS="$HOME/.dokeys"
BINARY_URL="https://raw.githubusercontent.com/jeffotoni/copyspace/refs/heads/master/v1/copyspace"

COLOR_GREEN='\033[0;32m'
COLOR_YELLOW='\033[1;33m'
COLOR_RESET='\033[0m'

# Detect OS (Linux or Mac)
OS="$(uname -s)"
if [[ "$OS" != "Linux" && "$OS" != "Darwin" ]]; then
    echo -e "${COLOR_YELLOW}Unsupported OS: $OS${COLOR_RESET}"
    exit 1
fi

# Check dependencies
if command -v wget >/dev/null 2>&1; then
    DL_CMD="wget -q -O"
elif command -v curl >/dev/null 2>&1; then
    DL_CMD="curl -fsSL -o"
else
    echo -e "${COLOR_YELLOW}Please install wget or curl before running this script.${COLOR_RESET}"
    exit 1
fi

# Ensure install dir exists and is writable
if [ ! -d "$INSTALL_DIR" ]; then
    sudo mkdir -p "$INSTALL_DIR"
fi

if [ ! -w "$INSTALL_DIR" ]; then
    echo -e "${COLOR_YELLOW}Root permission required to install binary. Please enter your password...${COLOR_RESET}"
    sudo chown "$USER" "$INSTALL_DIR"
fi

# Download the binary
echo -e "${COLOR_GREEN}Downloading $EXEC binary to $BIN_PATH...${COLOR_RESET}"
$DL_CMD "$BIN_PATH" "$BINARY_URL"

chmod +x "$BIN_PATH"

# Create .dokeys config if not exists
if [ ! -f "$DOKEYS" ]; then
    cat <<EOF > "$DOKEYS"
{
  "key": "key-digitalocean",
  "secret": "secret-digitalocean",
  "endpoint": "https://your-space.digitaloceanspaces.com",
  "region": "us-east-1",
  "bucket": "your-bucket-default"
}
EOF
    echo -e "${COLOR_GREEN}Created example config at $DOKEYS${COLOR_RESET}"
else
    echo -e "${COLOR_GREEN}Config $DOKEYS found. Skipping creation.${COLOR_RESET}"
fi

# Reload shell profile if needed
if [[ -f "$HOME/.zshrc" ]]; then
    SHELLRC="$HOME/.zshrc"
elif [[ -f "$HOME/.bashrc" ]]; then
    SHELLRC="$HOME/.bashrc"
else
    SHELLRC=""
fi

echo -e "${COLOR_GREEN}#########################################################${COLOR_RESET}"
echo -e "${COLOR_YELLOW}Copyspace successfully installed to $BIN_PATH!${COLOR_RESET}"
echo -e "${COLOR_YELLOW}You just need to configure your ~/.dokeys file if needed.${COLOR_RESET}"
echo -e "${COLOR_YELLOW}To get started, run:${COLOR_RESET} ${COLOR_GREEN}copyspace -h${COLOR_RESET}"
echo
echo -e "${COLOR_GREEN}Sample usage:${COLOR_RESET}"
echo -e "${COLOR_YELLOW}  copyspace -file /path/to/file.txt -bucket bucket-name${COLOR_RESET}"
echo -e "${COLOR_YELLOW}  copyspace -file /path/to/dir -bucket bucket-name -worker 100${COLOR_RESET}"
echo -e "${COLOR_YELLOW}  copyspace -cp -bucket bucket-name -out /path/to/dest${COLOR_RESET}"

if [[ -n "$SHELLRC" ]]; then
    echo -e "${COLOR_GREEN}If you have issues running 'copyspace', restart your terminal or source your shell config:${COLOR_RESET}"
    echo -e "${COLOR_YELLOW}  source $SHELLRC${COLOR_RESET}"
fi

echo -e "${COLOR_GREEN}######### Thanks for Download ##########${COLOR_RESET}"