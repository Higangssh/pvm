#!/usr/bin/env sh
set -eu

REPO="Higangssh/pvm"
INSTALL_DIR="${HOME}/.local/bin"
BINARY="pvm"

case "$(uname -s)" in
  Darwin) OS="darwin" ;;
  *) echo "This installer currently supports macOS only." >&2; exit 1 ;;
esac

case "$(uname -m)" in
  arm64|aarch64) ARCH="arm64" ;;
  x86_64) ARCH="amd64" ;;
  *) echo "Unsupported architecture: $(uname -m)" >&2; exit 1 ;;
esac

ASSET="${BINARY}-${OS}-${ARCH}"
API_URL="https://api.github.com/repos/${REPO}/releases/latest"

printf '==> Installing pvm for %s/%s\n' "$OS" "$ARCH"
RELEASE_JSON=$(curl -fsSL "$API_URL")
TAG=$(printf '%s' "$RELEASE_JSON" | grep '"tag_name":' | head -1 | sed -E 's/.*"([^"]+)".*/\1/')
URL=$(RELEASE_JSON="$RELEASE_JSON" python3 - "$ASSET" <<'PY'
import json, os, sys
asset = sys.argv[1]
data = json.loads(os.environ['RELEASE_JSON'])
for item in data.get('assets', []):
    if item.get('name') == asset:
        print(item.get('browser_download_url', ''))
        break
PY
)

if [ -z "$URL" ]; then
  echo "Asset not found for $ASSET in release $TAG" >&2
  exit 1
fi

mkdir -p "$INSTALL_DIR"
TARGET="$INSTALL_DIR/$BINARY"
curl -fsSL "$URL" -o "$TARGET"
chmod +x "$TARGET"

echo "==> Installed to $TARGET"

case ":$PATH:" in
  *":$INSTALL_DIR:"*)
    echo "==> $INSTALL_DIR already in PATH"
    ;;
  *)
    SHELL_NAME=$(basename "${SHELL:-/bin/zsh}")
    case "$SHELL_NAME" in
      zsh) RC_FILE="$HOME/.zshrc" ;;
      bash) RC_FILE="$HOME/.bashrc" ;;
      *) RC_FILE="$HOME/.profile" ;;
    esac
    LINE='export PATH="$HOME/.local/bin:$PATH"'
    if [ ! -f "$RC_FILE" ] || ! grep -Fq "$LINE" "$RC_FILE"; then
      printf '\n%s\n' "$LINE" >> "$RC_FILE"
      echo "==> Added $INSTALL_DIR to PATH in $RC_FILE"
      echo "    Run: source $RC_FILE"
    fi
    ;;
esac

echo
echo "Installed successfully!"
echo "Try: pvm --help"
