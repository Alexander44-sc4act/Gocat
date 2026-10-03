#!/bin/bash
# GoCat Man Page Installation Script

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Man page directory
MAN_DIR="/usr/local/share/man/man1"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DOCS_DIR="$(dirname "$SCRIPT_DIR")/docs/man"

echo -e "${BLUE}GoCat Man Page Installer${NC}"
echo "========================="
echo ""

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    echo -e "${YELLOW}Note: Running without root privileges. Using sudo...${NC}"
    SUDO="sudo"
else
    SUDO=""
fi

# Create man directory if it doesn't exist
if [ ! -d "$MAN_DIR" ]; then
    echo -e "${YELLOW}Creating man directory: $MAN_DIR${NC}"
    $SUDO mkdir -p "$MAN_DIR"
fi

# Install man pages
echo -e "${BLUE}Installing man pages...${NC}"

MAN_PAGES=(
    "gocat.1"
    "gocat-connect.1"
    "gocat-listen.1"
    "gocat-scan.1"
    "gocat-sniffer.1"
    "gocat-proxy.1"
)

for page in "${MAN_PAGES[@]}"; do
    if [ -f "$DOCS_DIR/$page" ]; then
        echo "  Installing $page..."
        $SUDO cp "$DOCS_DIR/$page" "$MAN_DIR/"
        $SUDO gzip -f "$MAN_DIR/$page"
        echo -e "  ${GREEN}✓${NC} $page installed"
    else
        echo -e "  ${YELLOW}⚠${NC} $page not found, skipping"
    fi
done

# Update man database
echo ""
echo -e "${BLUE}Updating man database...${NC}"
if command -v mandb &> /dev/null; then
    $SUDO mandb -q
    echo -e "${GREEN}✓${NC} Man database updated"
elif command -v makewhatis &> /dev/null; then
    $SUDO makewhatis
    echo -e "${GREEN}✓${NC} Man database updated"
else
    echo -e "${YELLOW}⚠${NC} Could not update man database (mandb/makewhatis not found)"
fi

echo ""
echo -e "${GREEN}Installation complete!${NC}"
echo ""
echo "You can now view the man pages with:"
echo "  man gocat"
echo "  man gocat-connect"
echo "  man gocat-listen"
echo "  man gocat-scan"
echo "  man gocat-sniffer"
echo "  man gocat-proxy"
