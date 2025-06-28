#!/bin/bash

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

USER_HOME="${HOME:-/root}"
PROJECT_DIR="$USER_HOME/NUMParser"
SERVICE_NAME="numparser"
GO_PATH="/usr/local/go"
GO_IN_SHELL_CONFIG_LINE='export PATH=$PATH:/usr/local/go/bin'
NEED_RELOAD=false

function confirm {
    local prompt="$1"
    local default="${2:-n}"
    read -p "$prompt" response
    case "${response:-$default}" in
        [Yy]*) return 0 ;;
        *) return 1 ;;
    esac
}

function reload_shell {
    echo -e "${YELLOW}\nChanges to PATH require a shell reload.${NC}"
    if confirm "Would you like to reload the shell now? (Y/n) " "y"; then
        echo -e "${GREEN}Reloading shell...${NC}"
        cd ~ || cd /tmp
        exec $SHELL -l
    else
        echo -e "${YELLOW}Please manually reload your shell or run:${NC}"
        echo -e "  source $(get_shell_config)"
    fi
}

function get_shell_config {
    if [ -n "$ZSH_VERSION" ]; then
        if [ -f "$USER_HOME/.zprofile" ]; then
            echo "$USER_HOME/.zprofile"
        else
            echo "$USER_HOME/.zshrc"
        fi
    else
        if [ -f "$USER_HOME/.profile" ]; then
            echo "$USER_HOME/.profile"
        else
            echo "$USER_HOME/.bashrc"
        fi
    fi
}

function stop_and_disable_service {
    echo -e "${YELLOW}Stopping and disabling systemd service...${NC}"
    if systemctl list-unit-files | grep -q "$SERVICE_NAME.service"; then
        sudo systemctl stop "$SERVICE_NAME"
        sudo systemctl disable "$SERVICE_NAME"
        sudo rm -f "/etc/systemd/system/${SERVICE_NAME}.service"
        sudo systemctl daemon-reload
        echo -e "${GREEN}Service $SERVICE_NAME removed.${NC}"
    else
        echo -e "${YELLOW}Service $SERVICE_NAME not found.${NC}"
    fi
}

function remove_project {
    if [ -d "$PROJECT_DIR" ]; then
        echo -e "${YELLOW}Removing project directory: ${PROJECT_DIR}${NC}"
        rm -rf "$PROJECT_DIR"
        echo -e "${GREEN}Project directory removed.${NC}"
    else
        echo -e "${YELLOW}Project directory not found: ${PROJECT_DIR}${NC}"
    fi
}

function remove_go {
    if [ -d "$GO_PATH" ]; then
        echo -e "${YELLOW}Removing Go installation at ${GO_PATH}...${NC}"
        sudo rm -rf "$GO_PATH"
        echo -e "${GREEN}Go removed from /usr/local/go.${NC}"
    else
        echo -e "${YELLOW}Go not found at /usr/local/go.${NC}"
    fi

    # Remove Go from PATH in shell config
    local SHELL_CONFIG
    SHELL_CONFIG=$(get_shell_config)
    if grep -q "$GO_IN_SHELL_CONFIG_LINE" "$SHELL_CONFIG"; then
        sed -i "\|$GO_IN_SHELL_CONFIG_LINE|d" "$SHELL_CONFIG"
        echo -e "${GREEN}Removed Go path from $SHELL_CONFIG.${NC}"
        NEED_RELOAD=true
    fi
}

function main {
    echo -e "${GREEN}Starting NUMParser uninstallation...${NC}"

    stop_and_disable_service
    remove_project

    if confirm "Do you want to also remove Go? (y/N): " "n"; then
        remove_go
    else
        echo -e "${YELLOW}Go will be kept on the system.${NC}"
    fi

    echo -e "${GREEN}Uninstallation complete.${NC}"

    if $NEED_RELOAD; then
        reload_shell
    else
        echo -e "${YELLOW}You may need to open a new terminal session for changes to take effect.${NC}"
    fi
}

main