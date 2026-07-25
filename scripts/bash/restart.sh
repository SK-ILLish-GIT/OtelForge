#!/usr/bin/env bash
set -euo pipefail
sudo systemctl restart otelcol
sudo systemctl is-active otelcol
