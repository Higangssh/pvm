#!/usr/bin/env sh
set -eu

INSTALL_DIR="${HOME}/.local/bin"
TARGET="$INSTALL_DIR/pvm"
if [ -n "${XDG_CONFIG_HOME:-}" ]; then
  CONFIG_ROOT="$XDG_CONFIG_HOME"
else
  CONFIG_ROOT="$HOME/Library/Application Support"
fi
CONFIG_DIR="$CONFIG_ROOT/pvm"

printf '==> Uninstalling pvm...\n'

if [ -f "$TARGET" ]; then
  rm -f "$TARGET"
  echo "    Removed $TARGET"
else
  echo "    $TARGET not found (skipped)"
fi

if [ -t 0 ]; then
  printf 'Remove config at %s too? (y/N): ' "$CONFIG_DIR"
  read -r ans || true
else
  ans="n"
  echo "    Non-interactive mode: keeping $CONFIG_DIR"
fi

case "$ans" in
  y|Y|yes|YES)
    rm -rf "$CONFIG_DIR"
    echo "    Removed $CONFIG_DIR"
    ;;
  *)
    echo "    Kept $CONFIG_DIR"
    ;;
esac

echo
echo "Uninstalled. If you added ~/.local/bin to PATH manually, keep or remove that line from your shell rc as desired."
