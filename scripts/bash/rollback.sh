#!/usr/bin/env bash
set -euo pipefail
CONFIG_DEST="/etc/otelcol/config.yaml"
BACKUP="${CONFIG_DEST}.bak"
if [ ! -f "$BACKUP" ]; then
  echo "no backup found" >&2
  exit 1
fi
sudo cp "$BACKUP" "$CONFIG_DEST"
sudo otelcol validate --config="$CONFIG_DEST"
sudo systemctl restart otelcol
sudo systemctl is-active otelcol
