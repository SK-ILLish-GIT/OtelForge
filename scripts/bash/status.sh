#!/usr/bin/env bash
set -euo pipefail
sudo systemctl is-active otelcol || true
systemctl is-active otelcol 2>/dev/null || echo "inactive"
