#!/usr/bin/env bash
set -euo pipefail
CONFIG_SRC="$1"
sudo otelcol validate --config="$CONFIG_SRC"
