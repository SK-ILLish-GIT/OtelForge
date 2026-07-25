#!/usr/bin/env bash
set -euo pipefail
sudo journalctl -u otelcol --no-pager -n 200
